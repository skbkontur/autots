package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

func main() {
	var listen, upstream string
	flag.StringVar(&listen, "listen", ":8181", "Listen address")
	flag.StringVar(&upstream, "upstream", "http://localhost:8080", "Upstream address")
	flag.Parse()

	remote, err := url.Parse(upstream)
	if err != nil {
		log.Panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	http.HandleFunc("/", handler(proxy))
	err = http.ListenAndServe(listen, nil)
	if err != nil {
		log.Panic(err)
	}
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			log.Println("ERROR",http.StatusText(405))
			http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
			r.Body.Close()
			return
		}
		// No need to modify search requests
		if !strings.HasSuffix(r.RequestURI, "_search") {
			if err := modifyRequest(r); err != nil {
				log.Println("ERROR", err)
				http.Error(w, http.StatusText(400), http.StatusBadRequest)
				r.Body.Close()
				return
			}
		}
		p.ServeHTTP(w, r)

	}
}

func modifyRequest(r *http.Request) error {
	var message map[string]interface{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&message); err != nil {
		return err
	}
	oldBody, _ := json.Marshal(message)

	if _, ok := message["@timestamp"]; !ok {
		if _, ok := message["_timestamp"]; !ok {
			message["@timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
		} else {
			message["@timestamp"] = message["_timestamp"]
		}
	}

	delete(message, "_timestamp")
	newBody, _ := json.Marshal(message)

	log.Println(string(oldBody), "->", string(newBody))
	r.Body = ioutil.NopCloser(bytes.NewReader(newBody))
	r.ContentLength = int64(len(newBody))
	return nil
}
