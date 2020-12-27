package pcopy

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

const (
	managerTickerInterval = 30 * time.Second
	defaultMaxAuthAge     = time.Minute
	noAuthRequestAge      = 0
	certCommonName        = "pcopy"
)

var (
	hmacAuthFormat        = "HMAC %d %d %s" // timestamp ttl b64-hmac
	hmacAuthRegex         = regexp.MustCompile(`^HMAC (\d+) (\d+) (.+)$`)
	hmacAuthOverrideParam = "a"
	clipboardRegex        = regexp.MustCompile(`^/c(?:/([-_a-zA-Z0-9]{1,100}))$`)
	clipboardPathFormat   = "/c/%s"
	clipboardDefaultPath  = "/c"

	//go:embed "web/index.gohtml"
	webTemplateSource string
	webTemplate       = template.Must(template.New("index").Funcs(templateFnMap).Parse(webTemplateSource))

	//go:embed web/static
	webStaticFs embed.FS

	//go:embed "scripts/join.sh.tmpl"
	joinTemplateSource string
	joinTemplate       = template.Must(template.New("join").Funcs(templateFnMap).Parse(joinTemplateSource))

	//go:embed "scripts/install.sh.tmpl"
	installTemplateSource string
	installTemplate       = template.Must(template.New("install").Funcs(templateFnMap).Parse(installTemplateSource))
)

// infoResponse is the response returned by the / endpoint
type infoResponse struct {
	ServerAddr string `json:"serverAddr"`
	Salt       string `json:"salt"`
}

type server struct {
	config       *Config
	countLimiter *limiter
	sizeLimiter  *limiter
	sync.Mutex
}

func Serve(config *Config) error {
	if err := checkConfig(config); err != nil {
		return err
	}
	server := &server{
		config:       config,
		sizeLimiter:  newLimiter(config.ClipboardSizeLimit),
		countLimiter: newLimiter(int64(config.FileCountLimit)),
	}
	go server.clipboardManager()
	return server.listenAndServeTLS()
}

func checkConfig(config *Config) error {
	if config.ListenAddr == "" {
		return errListenAddrMissing
	}
	if config.KeyFile == "" {
		return errKeyFileMissing
	}
	if config.CertFile == "" {
		return errCertFileMissing
	}
	if unix.Access(config.ClipboardDir, unix.W_OK) != nil {
		return errClipboardDirNotWritable
	}
	return nil
}

func (s *server) listenAndServeTLS() error {
	http.HandleFunc("/info", s.handleInfo)
	http.HandleFunc("/verify", s.handleVerify)
	http.HandleFunc("/install", s.handleInstall)
	http.HandleFunc("/join", s.handleJoin)
	http.HandleFunc("/download", s.handleDownload)
	http.HandleFunc("/c/", s.handleClipboard)
	http.HandleFunc("/c", s.handleClipboard)

	if s.config.WebUI {
		http.HandleFunc("/", s.handleWebRoot)
	}

	return http.ListenAndServeTLS(s.config.ListenAddr, s.config.CertFile, s.config.KeyFile, nil)
}

func (s *server) handleInfo(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s - %s %s", r.RemoteAddr, r.Method, r.RequestURI)

	salt := ""
	if s.config.Key != nil {
		salt = base64.StdEncoding.EncodeToString(s.config.Key.Salt)
	}

	response := &infoResponse{
		ServerAddr: s.config.ServerAddr,
		Salt:       salt,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.fail(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (s *server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if err := s.authorize(r); err != nil {
		s.fail(w, r, http.StatusUnauthorized, err)
		return
	}

	log.Printf("%s - %s %s", r.RemoteAddr, r.Method, r.RequestURI)
}

func (s *server) handleWebRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		if err := webTemplate.Execute(w, s.config); err != nil {
			s.fail(w, r, http.StatusInternalServerError, err)
		}
	} else if strings.HasPrefix(r.URL.Path, "/static") {
		r.URL.Path = "/web" + r.URL.Path // This is a hack to get the embedded path
		http.FileServer(http.FS(webStaticFs)).ServeHTTP(w, r)
	}
}

func (s *server) handleClipboard(w http.ResponseWriter, r *http.Request) {
	if err := s.authorize(r); err != nil {
		s.fail(w, r, http.StatusUnauthorized, err)
		return
	}

	log.Printf("%s - %s %s", r.RemoteAddr, r.Method, r.RequestURI)

	if err := os.MkdirAll(s.config.ClipboardDir, 0700); err != nil {
		s.fail(w, r, http.StatusInternalServerError, err)
		return
	}

	var id string
	matches := clipboardRegex.FindStringSubmatch(r.URL.Path)
	if matches == nil {
		id = DefaultId
	} else {
		id = matches[1]
	}
	file := fmt.Sprintf("%s/%s", s.config.ClipboardDir, id)

	if r.Method == http.MethodGet {
		s.handleClipboardGet(w, r, file)
	} else if r.Method == http.MethodPut {
		s.handleClipboardPut(w, r, file)
	}
}

func (s *server) handleClipboardGet(w http.ResponseWriter, r *http.Request, file string) {
	stat, err := os.Stat(file)
	if err != nil {
		s.fail(w, r, http.StatusNotFound, err)
		return
	}
	w.Header().Set("Length", strconv.FormatInt(stat.Size(), 10))
	f, err := os.Open(file)
	if err != nil {
		s.fail(w, r, http.StatusNotFound, err)
		return
	}
	defer f.Close()

	if _, err = io.Copy(w, f); err != nil {
		s.fail(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (s *server) handleClipboardPut(w http.ResponseWriter, r *http.Request, file string) {
	// Check total file count limit (only if file didn't exist already)
	stat, _ := os.Stat(file)
	if stat == nil {
		if err := s.countLimiter.Add(1); err != nil {
			s.fail(w, r, http.StatusBadRequest, err)
			return
		}
	}

	// Create new file or truncate existing
	f, err := os.OpenFile(file, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0600)
	if err != nil {
		s.countLimiter.Sub(1)
		s.fail(w, r, http.StatusInternalServerError, err)
		return
	}
	defer f.Close()
	defer s.updateStatsAndExpire()

	// Handle empty body
	if r.Body == nil {
		return
	}

	// Copy file contents (with file limit & total limit)
	fileSizeLimiter := newLimiter(s.config.FileSizeLimit)
	limitWriter := newLimitWriter(f, fileSizeLimiter, s.sizeLimiter)

	if _, err := io.Copy(limitWriter, r.Body); err != nil {
		if err == limitReachedError {
			s.fail(w, r, http.StatusBadRequest, err)
		} else {
			s.fail(w, r, http.StatusInternalServerError, err)
		}
		os.Remove(file)
		return
	}
	if r.Body.Close() != nil {
		s.fail(w, r, http.StatusInternalServerError, err)
		os.Remove(file)
		return
	}
}

func (s *server) handleDownload(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s - %s %s", r.RemoteAddr, r.Method, r.RequestURI)

	executable, err := getExecutable()
	if err != nil {
		s.fail(w, r, http.StatusInternalServerError, err)
		return
	}

	f, err := os.Open(executable)
	if err != nil {
		s.fail(w, r, http.StatusInternalServerError, err)
		return
	}
	defer f.Close()

	if _, err = io.Copy(w, f); err != nil {
		s.fail(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (s *server) handleInstall(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s - %s %s", r.RemoteAddr, r.Method, r.RequestURI)

	if err := installTemplate.Execute(w, s.config); err != nil {
		s.fail(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (s *server) handleJoin(w http.ResponseWriter, r *http.Request) {
	if err := s.authorize(r); err != nil {
		s.fail(w, r, http.StatusUnauthorized, err)
		return
	}

	log.Printf("%s - %s %s", r.RemoteAddr, r.Method, r.RequestURI)

	if err := joinTemplate.Execute(w, s.config); err != nil {
		s.fail(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (s *server) authorize(r *http.Request) error {
	if s.config.Key == nil {
		return nil
	}

	auth := r.Header.Get("Authorization")
	if encodedQueryAuth, ok := r.URL.Query()[hmacAuthOverrideParam]; ok && len(encodedQueryAuth) > 0 {
		queryAuth, err := base64.StdEncoding.DecodeString(encodedQueryAuth[0])
		if err != nil {
			log.Printf("%s - %s %s - cannot decode query auth override", r.RemoteAddr, r.Method, r.RequestURI)
			return errInvalidAuth
		}
		auth = string(queryAuth)
	}

	matches := hmacAuthRegex.FindStringSubmatch(auth)
	if matches == nil {
		log.Printf("%s - %s %s - auth header missing", r.RemoteAddr, r.Method, r.RequestURI)
		return errInvalidAuth
	}

	timestamp, err := strconv.Atoi(matches[1])
	if err != nil {
		log.Printf("%s - %s %s - hmac timestamp conversion: %s", r.RemoteAddr, r.Method, r.RequestURI, err.Error())
		return errInvalidAuth
	}

	ttlSecs, err := strconv.Atoi(matches[2])
	if err != nil {
		log.Printf("%s - %s %s - hmac ttl conversion: %s", r.RemoteAddr, r.Method, r.RequestURI, err.Error())
		return errInvalidAuth
	}

	hash, err := base64.StdEncoding.DecodeString(matches[3])
	if err != nil {
		log.Printf("%s - %s %s - hmac base64 conversion: %s", r.RemoteAddr, r.Method, r.RequestURI, err.Error())
		return errInvalidAuth
	}

	// Recalculate HMAC
	data := []byte(fmt.Sprintf("%d:%d:%s:%s", timestamp, ttlSecs, r.Method, r.URL.Path))
	hm := hmac.New(sha256.New, s.config.Key.Bytes)
	if _, err := hm.Write(data); err != nil {
		log.Printf("%s - %s %s - hmac calculation: %s", r.RemoteAddr, r.Method, r.RequestURI, err.Error())
		return errInvalidAuth
	}
	rehash := hm.Sum(nil)

	// Compare HMAC in constant time (to prevent timing attacks)
	if subtle.ConstantTimeCompare(hash, rehash) != 1 {
		log.Printf("%s - %s %s - hmac invalid", r.RemoteAddr, r.Method, r.RequestURI)
		return errInvalidAuth
	}

	// Compare timestamp (to prevent replay attacks)
	maxAge := defaultMaxAuthAge
	if ttlSecs > 0 {
		maxAge = time.Second * time.Duration(ttlSecs)
	}
	if maxAge > 0 {
		age := time.Now().Sub(time.Unix(int64(timestamp), 0))
		if age > maxAge {
			log.Printf("%s - %s %s - hmac request age mismatch", r.RemoteAddr, r.Method, r.RequestURI)
			return errInvalidAuth
		}
	}

	return nil
}

func (s *server) clipboardManager() {
	ticker := time.NewTicker(managerTickerInterval)
	for {
		s.updateStatsAndExpire()
		<- ticker.C
	}
}

func (s *server) updateStatsAndExpire() {
	s.Lock()
	defer s.Unlock()
	files, err := ioutil.ReadDir(s.config.ClipboardDir)
	if err != nil {
		log.Printf("error reading clipboard: %s", err.Error())
		return
	}
	numFiles := int64(0)
	totalSize := int64(0)
	for _, f := range files {
		if !s.maybeExpire(f) {
			numFiles++
			totalSize += f.Size()
		}
	}
	s.countLimiter.Set(numFiles)
	s.sizeLimiter.Set(totalSize)
	s.printStats()
}

func (s *server) printStats() {
	var countLimit, sizeLimit string
	if s.countLimiter.Limit() == 0 {
		countLimit = "no limit"
	} else {
		countLimit = fmt.Sprintf("max %d", s.countLimiter.Limit())
	}
	if s.sizeLimiter.Limit() == 0 {
		sizeLimit = "no limit"
	} else {
		sizeLimit = fmt.Sprintf("max %s", BytesToHuman(s.sizeLimiter.Limit()))
	}
	log.Printf("files: %d (%s), size %s (%s)", s.countLimiter.Value(), countLimit,
		BytesToHuman(s.sizeLimiter.Value()), sizeLimit)
}

// maybeExpire deletes a file if it has expired and returns true if it did
func (s *server) maybeExpire(file os.FileInfo) bool {
	if s.config.FileExpireAfter == 0 || time.Now().Sub(file.ModTime()) <= s.config.FileExpireAfter {
		return false
	}
	if err := os.Remove(filepath.Join(s.config.ClipboardDir, file.Name())); err != nil {
		log.Printf("failed to remove clipboard entry after expiry: %s", err.Error())
	}
	log.Printf("removed expired entry %s (%s)", file.Name(), BytesToHuman(file.Size()))
	return true
}

func (s *server) fail(w http.ResponseWriter, r *http.Request, code int, err error) {
	log.Printf("%s - %s %s - %s", r.RemoteAddr, r.Method, r.RequestURI, err.Error())
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf("%d", code)))
}

func getExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	realpath, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}

	return realpath, nil
}

var errListenAddrMissing = errors.New("listen address missing, add 'ListenAddr' to config or pass -listen")
var errKeyFileMissing = errors.New("private key file missing, add 'KeyFile' to config or pass -keyfile")
var errCertFileMissing = errors.New("certificate file missing, add 'CertFile' to config or pass -certfile")
var errClipboardDirNotWritable = errors.New("clipboard dir not writable by user")
var errInvalidAuth = errors.New("invalid auth")
