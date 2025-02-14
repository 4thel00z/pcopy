package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/4thel00z/pcopy/clipboard/clipboardtest"
	"github.com/4thel00z/pcopy/config/configtest"
	"github.com/4thel00z/pcopy/test"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestCLI_Copy(t *testing.T) {
	filename, config := configtest.NewTestConfig(t)
	serverRouter := startTestServerRouter(t, config)
	defer serverRouter.Stop()

	test.WaitForPortUp(t, "12345")

	app, stdin, _, stderr := newTestApp()
	stdin.WriteString("test stdin")

	if err := Run(app, "pcp", "-c", filename); err != nil {
		t.Fatal(err)
	}

	clipboardtest.Content(t, config, "default", "test stdin")
	test.StrContains(t, stderr.String(), "Direct link (valid for 7d")
	test.StrContains(t, stderr.String(), "curl -sSLk --pinnedpubkey")
	test.StrContains(t, stderr.String(), "https://localhost:12345/default")
}

func TestCLI_CopyPaste(t *testing.T) {
	filename, config := configtest.NewTestConfig(t)
	serverRouter := startTestServerRouter(t, config)
	defer serverRouter.Stop()

	test.WaitForPortUp(t, "12345")

	copyApp, copyStdin, _, copyStderr := newTestApp()
	copyStdin.WriteString("this is a test string")
	if err := Run(copyApp, "pcp", "-c", filename, "somefile"); err != nil {
		t.Fatal(err)
	}
	pasteApp, _, pasteStdout, _ := newTestApp()
	if err := Run(pasteApp, "ppaste", "-c", filename, "somefile"); err != nil {
		t.Fatal(err)
	}

	test.StrContains(t, copyStderr.String(), "https://localhost:12345/somefile")
	test.StrContains(t, pasteStdout.String(), "this is a test string")
}

func TestCLI_CopyPasteStream(t *testing.T) {
	filename, config := configtest.NewTestConfig(t)
	serverRouter := startTestServerRouter(t, config)
	defer serverRouter.Stop()

	test.WaitForPortUp(t, "12345")

	// Copy
	copyApp, copyStdin, _, copyStderr := newTestApp()
	copyErrChan := make(chan error)
	go func() {
		copyStdin.WriteString("this is a test string\n")
		if err := Run(copyApp, "pcp", "--stream", "-c", filename, "mystream"); err != nil {
			copyErrChan <- err
			return
		}
		test.StrContains(t, copyStderr.String(), "https://localhost:12345/mystream")
	}()

	// Wait for pipe to be created
	success := false
	for i := 0; i < 20; i++ {
		stat, _ := os.Stat(filepath.Join(config.ClipboardDir, "mystream"))
		if stat != nil && stat.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {
			success = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !success {
		t.Fatalf("waiting for pipe timed out")
	}

	// Awkwardly check for copy error, since we cannot call t.Fatal in a goroutine
	select {
	case err := <-copyErrChan:
		t.Fatal(err)
	default:
	}

	// Paste
	pasteApp, _, pasteStdout, _ := newTestApp()
	if err := Run(pasteApp, "ppaste", "-c", filename, "mystream"); err != nil {
		t.Fatal(err)
	}

	test.StrContains(t, pasteStdout.String(), "this is a test string")
}

func TestCurl_CopyPOSTSuccess(t *testing.T) {
	_, config := configtest.NewTestConfig(t)
	serverRouter := startTestServerRouter(t, config)
	defer serverRouter.Stop()

	test.WaitForPortUp(t, "12345")

	var stdout bytes.Buffer
	cmd := exec.Command("curl", "-sSLk", "-dabc", fmt.Sprintf("%s/howdy?f=json", config.ServerAddr))
	cmd.Stdout = &stdout
	cmd.Run()

	clipboardtest.Content(t, config, "howdy", "abc")
	test.StrContains(t, stdout.String(), `"url":"https://localhost:12345/howdy"`) // json
}

func TestCurl_POSTGETRandomWithJsonFormat(t *testing.T) {
	_, config := configtest.NewTestConfig(t)
	serverRouter := startTestServerRouter(t, config)
	defer serverRouter.Stop()

	test.WaitForPortUp(t, "12345")

	var stdout bytes.Buffer
	cmdCurlPOST := exec.Command("curl", "-sSLk", "-dabc", fmt.Sprintf("%s?f=json", config.ServerAddr))
	cmdCurlPOST.Stdout = &stdout
	cmdCurlPOST.Run()

	var info map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &info)

	stdout.Reset()
	cmdCurlGET := exec.Command("sh", "-c", info["curl"].(string))
	cmdCurlGET.Stdout = &stdout
	cmdCurlGET.Run()

	test.StrEquals(t, stdout.String(), "abc")
}

func TestCurl_POSTGETRandomStreamWithJsonFormat(t *testing.T) {
	// This tests #46: curl POST with streaming and short payloads does not work (curl -dabc http://...?s=1)

	_, config := configtest.NewTestConfig(t)
	serverRouter := startTestServerRouter(t, config)
	defer serverRouter.Stop()

	test.WaitForPortUp(t, "12345")

	// Streaming enabled (s=1), note that "stdbuf -oL" is required to flush buffers after every line
	cmdCurlPOST := exec.Command("stdbuf", "-oL", "curl", "-sSLk", "-dabc", fmt.Sprintf("%s?s=1&f=json", config.ServerAddr))
	stdoutPipe, _ := cmdCurlPOST.StdoutPipe()
	cmdCurlPOST.Start()

	out := test.WaitForOutput(t, stdoutPipe, 1*time.Second, 100*time.Millisecond)
	var info map[string]interface{}
	json.Unmarshal([]byte(out), &info)

	fileID := info["file"].(string)
	curlGET := info["curl"].(string)

	file := filepath.Join(config.ClipboardDir, fileID)
	stat, _ := os.Stat(file)
	if stat.Mode()&os.ModeNamedPipe == 0 {
		t.Fatalf("expected %s to be a pipe, but it's not", file)
	}

	// Now GET it
	var stdout bytes.Buffer
	cmdCurlGET := exec.Command("sh", "-c", curlGET)
	cmdCurlGET.Stdout = &stdout
	cmdCurlGET.Run()

	test.StrEquals(t, stdout.String(), "abc")
	stat, _ = os.Stat(file)
	if stat != nil {
		t.Fatalf("expected %s to not exist anymore, but it does", file)
	}
}
