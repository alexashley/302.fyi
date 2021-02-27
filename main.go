package main

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"os"
)

var (
	//go:embed config.yaml
	configYaml []byte
	scrambled  = "WW91bmcgZnJ5IG9mIHRyZWFjaGVyeSE="
)

type redirect struct {
	Path string
	Url  string
}

type config struct {
	Redirects []redirect
}

func egg(w http.ResponseWriter, _ *http.Request) {
	overEasy, _ := base64.StdEncoding.DecodeString(scrambled)
	_, _ = w.Write(overEasy)
}

func kettle(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(418)
}

func handler(r *redirect) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Show the link value if any query params are set, e.g., ?show=true
		// Or if the path ends in a plus sign (bit.ly convention)
		if len(req.URL.RawQuery) > 0 || req.URL.Path[len(req.URL.Path)-1:] == "+" {
			res := `
<html>
<pre>%s</pre>
<a href="%s">Click me to follow the link</a>
`
			fmt.Fprintf(w, fmt.Sprintf(res, r.Url, r.Url))
			return
		}

		w.Header().Set("Location", r.Url)
		w.WriteHeader(302)
	}
}

func main() {
	log.Println("Wrangling yaml...")
	c := config{}
	if err := yaml.Unmarshal(configYaml, &c); err != nil {
		log.Fatal("Unable to load config", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "1234"
	}
	for i, _ := range c.Redirects {
		r := c.Redirects[i]
		http.HandleFunc(r.Path, handler(&r))
		http.HandleFunc(r.Path+"+", handler(&r))
	}

	log.Println("Cracking eggs...")
	http.HandleFunc("/egg", egg)

	log.Println("Putting on the kettle...")
	http.HandleFunc("/", kettle)

	log.Printf("Listening on port %s", port)
	_ = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
