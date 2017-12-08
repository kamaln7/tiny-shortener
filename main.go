package main

import (
	"bytes"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	fileMutex    sync.RWMutex
	length       = flag.Int("length", 3, "code length")
	key          = flag.String("key", "", "upload API Key")
	root         = flag.String("root", "", "root redirect")
	urls         = flag.String("urls", "/srv/www/urls/", "path to urls")
	listenAddr   = flag.String("listenAddr", "127.0.0.1:5556", "listen address")
	publicURL    = flag.String("url", "http://127.0.0.1:5556/", "path to public facing url")
	notFoundPath = flag.String("notFound", "", "404 file")
	notFoundHTML []byte
	letterRunes  = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	*publicURL = strings.TrimRight(*publicURL, "/") + "/"

	if *key == "" {
		log.Fatal(errors.New("please pass a secret key"))
	}

	if *notFoundPath != "" {
		var err error
		notFoundHTML, err = ioutil.ReadFile(*notFoundPath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("using default 404 page")
		notFoundHTML = []byte(`<div style="color: #333; font-family:sans-serif; font-size: 24px; text-align: center; padding-top: 15px;"><strong>404</strong> not found</div>`)
	}

	log.Printf("listening on %s\n", *listenAddr)
	http.HandleFunc("/", serve)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func serve(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// root redirect & upload handlers
	if path == "/" {
		switch r.Method {
		case "GET":
			if *root != "" {
				http.Redirect(w, r, *root, 301)
			} else {
				notfound(w, r)
			}
		case "POST":
			create(w, r)
		}

		return
	}

	redirect(w, r, path[1:])
}

func redirect(w http.ResponseWriter, r *http.Request, code string) {
	code = path.Base(code)

	fileMutex.RLock()
	url, err := ioutil.ReadFile(filepath.Join(*urls, code))
	fileMutex.RUnlock()
	if err != nil {
		notfound(w, r)
		return
	}

	http.Redirect(w, r, string(bytes.TrimSpace(url)), 301)
}

func create(w http.ResponseWriter, r *http.Request) {
	var (
		url    = r.FormValue("url")
		apiKey = r.FormValue("key")
	)
	if url == "" || apiKey == "" || apiKey != *key {
		notfound(w, r)
		return
	}

	// find a nonexistent code
	code := r.FormValue("code")
	if code == "" {
		for {
			code = randomString(*length)
			if !codeExists(code) {
				break
			}
		}
	} else {
		if codeExists(code) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("code already exists"))
			return
		}
	}

	fileMutex.Lock()
	err := ioutil.WriteFile(filepath.Join(*urls, code), bytes.TrimSpace([]byte(url)), 0644)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fileMutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(*publicURL + code))
}

func notfound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	w.Write(notFoundHTML)
}

func codeExists(code string) bool {
	fileMutex.RLock()
	defer fileMutex.RUnlock()

	_, err := os.Stat(filepath.Join(*urls, path.Base(code)))
	return !os.IsNotExist(err)
}
