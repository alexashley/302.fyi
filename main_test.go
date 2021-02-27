package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	client          *http.Client
	conf            *config
	appUrl          string
	expectedVersion string
)

func TestMain(m *testing.M) {
	conf = &config{}
	if err := yaml.Unmarshal(configYaml, conf); err != nil {
		fmt.Println("Error unmarshalling config", err)
		os.Exit(1)
	}
	client = &http.Client{
		// don't follow redirects
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Second * 5,
	}

	rand.Seed(time.Now().Unix())

	appUrl = os.Getenv("APP_URL")
	if appUrl == "" {
		appUrl = "http://localhost:1234"
	}

	expectedVersion = os.Getenv("VERSION")
	if expectedVersion == "" {
		expectedVersion = "local"
	}

	os.Exit(m.Run())
}

func TestEgg(t *testing.T) {
	t.Parallel()
	r := makeRequest(t, "/egg")
	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.Equal(t, "Young fry of treachery!", readBody(r))
}

func TestTeapot(t *testing.T) {
	t.Parallel()
	r := makeRequest(t, "/not-a-real-path")
	assert.Equal(t, http.StatusTeapot, r.StatusCode)
}

func TestHealthz(t *testing.T) {
	t.Parallel()
	r := makeRequest(t, "/healthz")
	assert.Equal(t, http.StatusOK, r.StatusCode)
}

func TestRedirects(t *testing.T) {
	for _, redirect := range conf.Redirects {
		redirect := redirect // capture variable
		t.Run(fmt.Sprintf("%s should return %s", redirect.Path, redirect.Url), func(t *testing.T) {
			t.Parallel()

			r := makeRequest(t, redirect.Path)
			assert.Equal(t, 302, r.StatusCode)
			assert.Equal(t, redirect.Url, r.Header.Get("Location"))
		})
	}
}

func TestRevealUrl(t *testing.T) {
	for _, tc := range []string{"?reveal", "+"} {
		tc := tc
		t.Run("reveal url when path ends in "+tc, func(t *testing.T) {
			t.Parallel()
			randomRedirect := conf.Redirects[rand.Intn(len(conf.Redirects))]

			r := makeRequest(t, randomRedirect.Path+tc)
			body := readBody(r)

			assert.Equal(t, http.StatusOK, r.StatusCode)
			assert.True(t, strings.Contains(body, randomRedirect.Url))
		})
	}
}

func TestVersion(t *testing.T) {
	t.Parallel()

	r := makeRequest(t, "/version")

	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.Equal(t, expectedVersion, strings.TrimSpace(readBody(r)))
}

func makeRequest(t *testing.T, path string) *http.Response {
	r, err := client.Get(appUrl + path)
	assert.Nil(t, err)

	return r
}

func readBody(r *http.Response) string {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Sprint("Unable to read body: ", err)
	}

	return string(body)
}
