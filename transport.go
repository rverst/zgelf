package zgelf

import (
	"net/http"
	"time"
)

type TransportMode string

const (
	TransportHttp = TransportMode("http")
	TransportTcp  = TransportMode("tcp")
	TransportUdp  = TransportMode("udp")
)

type transport interface {
	Mode() TransportMode
	ServerUrl() string
	ServerPort() uint16
	BufferSize() int
	BufferTime() time.Duration
	SendBuffer(buffer *logBuffer) error
}

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

func (t *httpTransport) BufferTime() time.Duration {
	return t.duration
}

func (t *httpTransport) SendBuffer(buffer *logBuffer) error {

	if t.client == nil {
		t.client = &http.Client{
			Timeout: time.Second * 15,
		}
	}

	//req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d", t.url, t.port), buffer)

	return nil
}
