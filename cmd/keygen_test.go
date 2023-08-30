package cmd

import (
	"github.com/4thel00z/pcopy/pcopy/crypto"
	"github.com/4thel00z/pcopy/pcopy/test"
	"strings"
	"testing"
)

func TestCLI_Keygen(t *testing.T) {
	app, stdin, stdout, _ := newTestApp()
	stdin.WriteString("this is my password\nthis is my password")

	if err := Run(app, "pcopy", "keygen"); err != nil {
		t.Fatal(err)
	}

	line, _ := stdout.ReadString('\n')
	parts := strings.Split(strings.TrimSpace(line), " ")
	encodedKey := parts[1]
	test.StrEquals(t, "Key", parts[0])

	key, _ := crypto.DecodeKey(encodedKey)
	derivedKey := crypto.DeriveKey([]byte("this is my password"), key.Salt)

	test.BytesEquals(t, key.Salt, derivedKey.Salt)
	test.BytesEquals(t, key.Bytes, derivedKey.Bytes)
}
