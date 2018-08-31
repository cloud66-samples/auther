package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	backendHost  string
	backendPort  int
	frontendHost string
	frontendPort int
)

func NewResponse(r *http.Request, contentType string, status int, body string) *http.Response {
	resp := &http.Response{}
	resp.Request = r
	resp.TransferEncoding = r.TransferEncoding
	resp.Header = make(http.Header)
	resp.Header.Add("Content-Type", contentType)
	resp.StatusCode = status
	buf := bytes.NewBufferString(body)
	resp.ContentLength = int64(buf.Len())
	resp.Body = ioutil.NopCloser(buf)
	return resp
}

type AuthTransport struct {
	DelegateRoundTripper http.RoundTripper
}

func (t *AuthTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	user, pass, ok := req.BasicAuth()

	if ok && user == "user" && pass == "password" {
		log.Println("Authenticated")
		return t.DelegateRoundTripper.RoundTrip(req)
	} else {
		log.Println("Authenticating")
		resp = NewResponse(req, "text/plain", http.StatusUnauthorized, "")
		resp.Header.Add("WWW-Authenticate", `Basic realm="Cloud 66 Auther"`)
		resp.StatusCode = http.StatusUnauthorized
	}

	return resp, nil
}

func NewProxy(target *url.URL) *httputil.ReverseProxy {
	log.Printf("Target: %s \n", target.RequestURI())
	director := func(req *http.Request) {
		log.Printf("Request: [%s] %s\n", req.Method, req.URL)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
	}

	transport := &AuthTransport{http.DefaultTransport}

	return &httputil.ReverseProxy{Director: director, Transport: transport}
}

func main() {
	flag.StringVar(&frontendHost, "binding", "0.0.0.0", "Server listen address")
	flag.IntVar(&frontendPort, "port", 9090, "port")
	flag.StringVar(&backendHost, "backend-host", "127.0.0.1", "backend host")
	flag.IntVar(&backendPort, "backend-port", 5000, "backend port")
	flag.Parse()

	log.Printf("Backend: %s:%d\n", backendHost, backendPort)
	log.Printf("Frontend: %s:%d\n", frontendHost, frontendPort)
	u, _ := url.Parse(fmt.Sprintf("http://%s:%d", backendHost, backendPort))
	proxy := NewProxy(u)
	http.Handle("/scripts.js", proxy)
	http.Handle("/style.css", proxy)
	http.Handle("/background.jpg", proxy)
	http.Handle("/", proxy)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", frontendHost, frontendPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
