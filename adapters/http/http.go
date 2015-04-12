package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

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

	for m := range logstream {
		//		message := bytes.NewBufferString(message.Data)

		// Create a JSON representation of the log message
		httpMessage := HTTPMessage{
			Message:  m.Data,
			Time:     m.Time.Format(time.ISO8601),
			Source:   m.Source,
			Name:     m.Container.Name,
			ID:       m.Container.ID,
			Image:    m.Container.Config.Image,
			Hostname: m.Container.Config.Hostname,
		}
		message, err := json.Marshal(httpMessage)

		req, err := http.NewRequest("POST", a.url, bytes.NewReader(message))
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

// HTTPMessage is a simple JSON representation of the log message.
type HTTPMessage struct {
	Message  string `json:"message"`
	Time     string `json:"time"`
	Source   string `json:"source"`
	Name     string `json:"docker.name"`
	ID       string `json:"docker.id"`
	Image    string `json:"docker.image"`
	Hostname string `json:"docker.hostname"`
}
