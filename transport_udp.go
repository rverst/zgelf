package zgelf

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

const chunkSize = 1420
const chunkHeader = 12
const maxDataSize = chunkSize - chunkHeader // the maximum datagram size per chunk, should be less than the MTU

type UdpTransport struct {
	serverAddr *net.UDPAddr
	localAddr  *net.UDPAddr
	compress   bool
}

func NewUdpTransport(conn string) (*UdpTransport, error) {

	srvAddr, err := net.ResolveUDPAddr("udp", conn)
	if err != nil {
		return nil, err
	}

	locAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, err
	}

	t := UdpTransport{
		serverAddr: srvAddr,
		localAddr:  locAddr,
	}
	return &t, nil
}

func (t *UdpTransport) Mode() TransportMode {
	return TransportUdp
}

func (t *UdpTransport) BufferSize() int {
	// we do not buffer in udp mode, but since we are processing the
	// log packages, we still use the asynchronous handling
	return -1
}

func (t *UdpTransport) BufferTime() time.Duration {
	return 0
}

func (t *UdpTransport) SendBuffer(buffer *logBuffer) error {
	conn, err := net.DialUDP("udp", t.localAddr, t.serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	for buffer.Size() > 0 {
		d, err := buffer.Pull()
		if err != nil {
			return err
		}
		if len(d) <= maxDataSize {
			if _, err := conn.Write(d); err != nil {
				return err
			}
		} else {
			chunks := (len(d) / maxDataSize) + 1
			if chunks > 128 {
				return fmt.Errorf("buffer to big, exceeding maximum of 128 chunks: %d", chunks)
			}

			header := make([]byte, chunkHeader)
			rand.Read(header)
			header[0] = 0x1e
			header[1] = 0x0f
			header[11] = byte(chunks)

			for i := byte(0); i < byte(chunks); i++ {
				cData := make([]byte, chunkHeader)
				header[10] = i
				copy(cData, header)
				o := int(i) * maxDataSize
				r := len(d) - o

				if r > maxDataSize {
					cData = append(cData, d[o:o+maxDataSize]...)
				} else if r > 0 {
					cData = append(cData, d[o:o+r]...)
				} else {
					continue
				}
				if _, err := conn.Write(cData); err != nil {
					fmt.Printf("error sending chunk %v", err)
				}
			}
		}
	}

	return nil
}
