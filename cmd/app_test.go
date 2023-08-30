package cmd

import (
	"bytes"
	"github.com/urfave/cli/v2"
	"github.com/4thel00z/pcopy/pcopy/config"
	"github.com/4thel00z/pcopy/pcopy/server"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

// This only contains helpers so far

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func newTestApp() (*cli.App, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	var stdin, stdout, stderr bytes.Buffer
	app := New()
	app.Reader = &stdin
	app.Writer = &stdout
	app.ErrWriter = &stderr
	return app, &stdin, &stdout, &stderr
}

func startTestServerRouter(t *testing.T, config *config.Config) *server.Router {
	router, err := server.NewRouter(config)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		if err := router.Start(); err != nil && err != http.ErrServerClosed {
			panic(err) // 'go vet' complains about 't.Fatal(err)'
		}
	}()
	return router
}
