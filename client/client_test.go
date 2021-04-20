package client_test

import (
	"context"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/hamfist/k8s-http-auth/client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	td, err := ioutil.TempDir("", "k8s-http-auth.*")
	assert.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		_ = os.RemoveAll(td)
		cancel()
	}()

	tokenPath := filepath.Join(td, "api-token")

	clOpts := &client.Options{
		IDHeader:    "Whats-The-Password",
		TokenExpiry: 10 * time.Minute,
		TokenPath:   tokenPath,
	}
	cl := client.New(ctx, clOpts)

	log := logrusr.NewLogger(logrus.New()).WithName("k8s-http-auth-test")
	ctx = logr.NewContext(ctx, log)

	time.Sleep(100 * time.Millisecond)

	tok, err := cl.ID(ctx)
	assert.NotNilf(t, err, "initial read fails because the file %q does not exist", tokenPath)

	err = ioutil.WriteFile(tokenPath, []byte("sturgeon\n"), 0640)
	assert.Nilf(t, err, "writing to token path %q resulted in an error", tokenPath)

	time.Sleep(200 * time.Millisecond)

	tok, err = cl.ID(ctx)
	assert.Nilf(t, err, "subsequent read succeeds because the file %q exists now", tokenPath)
	assert.Equalf(t, "sturgeon", tok, "token matched expected value")

	req := httptest.NewRequest("GET", "/haddock", nil)
	req, err = cl.WithHeader(req)
	assert.Nilf(t, err, "request with header produced no error")
	assert.NotEqual(t, "", req.Header.Get(clOpts.IDHeader))
}
