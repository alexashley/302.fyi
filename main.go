package main

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"gopkg.in/yaml.v2"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
)

var (
	//go:embed config.yaml
	configYaml    []byte
	scrambled     = "WW91bmcgZnJ5IG9mIHRyZWFjaGVyeSE="
	reservedPaths = []string{"/healthz", "/version", "/egg", "/teapot"}
	//go:embed version
	versionFile string
)

type redirect struct {
	Path string
	Url  string
}

type config struct {
	Redirects []redirect
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, "ok")
}

func egg(w http.ResponseWriter, _ *http.Request) {
	overEasy, _ := base64.StdEncoding.DecodeString(scrambled)
	_, _ = w.Write(overEasy)
}

func kettle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func index(c *config) http.HandlerFunc {
	html := `
<html>
<h1>302.fyi</h1>
<p>Try one of these:</p>
<ul>
{{- range .Redirects -}}
<li>
	<a href="{{ .Path }}">{{ .Path }}</a> (<a href="{{ .Path }}+">+</a>)
</li>
{{ end }}
</ul>
<p>To see the URL before being redirected, click the + next to each link</p>
`
	t := template.Must(template.New("index").Parse(html))

	return func(w http.ResponseWriter, r *http.Request) {
		_ = t.ExecuteTemplate(w, "index", c)
	}
}

func version(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(versionFile))
}

func handler(r *redirect) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Show the link value if any query params are set, e.g., ?show=true
		// Or if the path ends in a plus sign (bit.ly convention)
		if len(req.URL.RawQuery) > 0 || req.URL.Path[len(req.URL.Path)-1:] == "+" {
			res := `
<html>
<pre>%s</pre>
<a href="%s">Click here to follow the link</a>
`
			_, _ = fmt.Fprintf(w, fmt.Sprintf(res, r.Url, r.Url))
			return
		}

		w.Header().Set("Location", r.Url)
		w.WriteHeader(http.StatusFound)
	}
}

func validateConfig(c *config) {
	seenPaths := map[string]string{}
	for i := range c.Redirects {
		r := c.Redirects[i]
		if _, ok := seenPaths[r.Path]; ok {
			log.Fatalf("Path %s is duplicated, mapped to %s and %s", r.Path, seenPaths[r.Path], r.Url)
		}

		if _, err := url.ParseRequestURI(r.Url); err != nil {
			log.Fatalf("%s is mapped to an invalid url: %s", r.Path, err)
		}

		for _, reserved := range reservedPaths {
			if r.Path == reserved {
				log.Fatalf("%s is a reserved path", reserved)
			}
		}

		seenPaths[r.Path] = r.Url
	}
}

func main() {
	log.Println("Wrangling yaml...")
	c := config{}
	if err := yaml.Unmarshal(configYaml, &c); err != nil {
		log.Fatal("Unable to load config", err)
	}
	validateConfig(&c)

	port := os.Getenv("PORT")
	if port == "" {
		port = "1234"
	}

	for i := range c.Redirects {
		r := c.Redirects[i]

		http.HandleFunc(r.Path, handler(&r))
		http.HandleFunc(r.Path+"+", handler(&r))
	}

	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/version", version)

	log.Println("Cracking eggs...")
	http.HandleFunc("/egg", egg)

	log.Println("Putting on the kettle...")
	http.HandleFunc("/teapot", kettle)

	http.HandleFunc("/", index(&c))

	log.Printf("Listening on port %s", port)
	_ = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
