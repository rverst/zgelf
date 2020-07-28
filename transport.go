package zgelf

import (
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
	BufferSize() int
	BufferTime() time.Duration
	SendBuffer(buffer *logBuffer) error
}

