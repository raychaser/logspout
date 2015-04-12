package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gliderlabs/logspout/router"
)

func init() {
	router.AdapterFactories.Register(NewHTTPAdapter, "http")
	router.AdapterFactories.Register(NewHTTPAdapter, "https")
}

func debug(v ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		log.Println(v...)
	}
}

func getopt(name, dfault string) string {
	value := os.Getenv(name)
	if value == "" {
		value = dfault
	}
	return value
}

// HTTPAdapter is an adapter that POSTs logs to an HTTP endpoint
type HTTPAdapter struct {
	route  *router.Route
	url    string
	client *http.Client
}

// NewHTTPAdapter creates an HTTPAdapter.
func NewHTTPAdapter(route *router.Route) (router.LogAdapter, error) {

	path := getopt("HTTP_PATH", "")
	url := fmt.Sprintf("%s://%s%s", route.Adapter, route.Address, path)
	debug("http adapter url:", url)

	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	return &HTTPAdapter{
		route:  route,
		url:    url,
		client: client,
	}, nil
}

// Stream implements the router.LogAdapter interface.
func (a *HTTPAdapter) Stream(logstream chan *router.Message) {

	for message := range logstream {
		message := bytes.NewBufferString(message.Data)
		req, err := http.NewRequest("POST", a.url, message)
		if err != nil {
			log.Println("syslog:", err)
			return
		}
		resp, err := a.client.Do(req)
		if err != nil {
			log.Println("syslog:", err)
			return
		}
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}
}
