package zgelf

import (
	"errors"
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

func NewHttpTransport(url string, port uint16) (*httpTransport, error) {

	return nil, errors.New("not implemented")

	//t := httpTransport{
	//	url:        url,
	//	port:       port,
	//	bufferSize: 64 * 1024,
	//	duration:   time.Second * 30,
	//}
	//return &t
}

func (t *httpTransport) Mode() TransportMode {
	return TransportHttp
}

func (t *httpTransport) BufferSize() int {
	return t.bufferSize
}
func (t *httpTransport) SendBuffer(buffer *logBuffer) error {

	return errors.New("not implemented")
}
