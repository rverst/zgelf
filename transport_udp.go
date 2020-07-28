package zgelf

import (
	"net/http"
	"time"
)

type httpTransport struct {
	url        string
	port       uint16
	bufferSize int
	duration   time.Duration
	client     *http.Client
}

func NewHttpTransport(url string, port uint16) *httpTransport {
	t := httpTransport{
		url:        url,
		port:       port,
		bufferSize: 64 * 1024,
		duration:   time.Second * 30,
	}
	return &t
}

func (t *httpTransport) Mode() TransportMode {
	return TransportHttp
}

func (t *httpTransport) ServerUrl() string {
	return t.url
}

func (t *httpTransport) ServerPort() uint16 {
	return t.port
}

func (t *httpTransport) BufferSize() int {
	return t.bufferSize
}