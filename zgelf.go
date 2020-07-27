package zgelf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	tempLogFileRegex       = "^log_([[:digit:]]+).[gz|log]$"
	defaultBufferSize      = 1024 * 64        // default buffer size (flush if size exceeds 60kB)
	defaultBufferTime      = time.Second * 30 // default buffer time (flush every 30 seconds)
	GelfVersion            = "1.1"
	ErrorFieldName         = "_err"
	ErrorStackFieldName    = "_err_stack"
	HostFieldName          = "host"
	FileFieldName          = "_file"
	FullMessageFieldName   = "full_message"
	LevelFieldName         = "level"
	OriginalLevelFieldName = "_o_level"
	TimestampFieldName     = "timestamp"
	ShortMessageFieldName  = "short_message"
	VersionFieldName       = "version"
	LineNumberFieldName    = "_line"
	NotAllowedIdFieldName  = "_id"
)

var ErrorKeyNotAllowed = errors.New("key `id` is not allowed")

type GelfWriter struct {
	transport   transport
	tempLogPath string
	host        string
	queue       chan map[string]interface{}
	wg          sync.WaitGroup
	mu          sync.RWMutex
	buffer      *logBuffer
	ticker      *time.Ticker
}

// New crates a new GelfWriter which can be used as a sink
// for zerolog. The parameter `host` is set as the appropriate field
// in the GELF-package, the server ist configured with tha parameters
// `serverUrl` and `serverPort` and mode. Transport over http(s) is the default.
func New(host, tmpLogPath string, trans transport) GelfWriter {
	w := GelfWriter{
		transport:   trans,
		tempLogPath: tmpLogPath,
		host:        host,
		buffer:      NewLogBuffer(),
		ticker:      time.NewTicker(defaultBufferTime),
		queue:       make(chan map[string]interface{}, 500),
	}

	go func() {
		for range w.ticker.C {
			w.Flush(true)
		}
	}()

	go w.worker()
	return w
}

func (w GelfWriter) Write(p []byte) (n int, err error) {

	var evt map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()
	err = d.Decode(&evt)
	if err != nil {
		return n, fmt.Errorf("cannot decode event: %s", err)
	}

	// ignore logs without message, since they are no allowed in GELF
	val, ok := evt[zerolog.MessageFieldName]
	if !ok || val == "" {
		return len(p), nil
	}
	l := len(w.queue)
	if l > 0 && l%50 == 0 {
		fmt.Printf("###queue: %d\n", l)
	}

	w.queue <- evt
	return len(p), nil
}

// Close waits for the queue to empty and
// all currently processed log entries to be finished,
// finally flushes the buffer
func (w *GelfWriter) Close() {
	time.Sleep(time.Millisecond * 10)

	// wait for queue to empty
	for len(w.queue) > 0 {
	}
	close(w.queue)

	// wait for process routines to finish
	w.wg.Wait()

	// flush buffer
	w.Flush(true)
	if !(w.buffer.Size() > 0) {
		fmt.Println("BUFFER NOT EMPTY")
	}
}

// Flush flushes the send buffer
func (w *GelfWriter) Flush(block bool) {

	if w.buffer.Size() == 0 {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	c := w.buffer.Copy()
	w.buffer.Clear()

	if block {
		_ = w.sendBuffer(c)
	} else {
		go func() {
			if err := w.sendBuffer(c); err == nil {
				w.sendTemporaryLogs()
			}
		}()
	}
}

// SetMaxBufferTime sets the time after the log-buffer is flushed,
// regardless of its size.
func (w *GelfWriter) SetMaxBufferTime(bufferTime time.Duration) {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	w.ticker = time.NewTicker(bufferTime)
}

func (w *GelfWriter) worker() {

	for data := range w.queue {
		w.wg.Add(1)
		go func(evt map[string]interface{}) {
			defer w.wg.Done()
			w.process(evt)
		}(data)
	}
}

func (w *GelfWriter) process(evt map[string]interface{}) {

	evn := make(map[string]interface{}, len(evt))
	for k, v := range evt {
		switch k {
		case zerolog.LevelFieldName:
			lvl := parseLogLevel(v.(string))
			if lvl < 0 {
				return
			}
			evn[LevelFieldName] = lvl
			evn[OriginalLevelFieldName] = v
		case zerolog.TimestampFieldName:
			t, err := convertTime(v.(json.Number), zerolog.TimeFieldFormat)
			if err == nil {
				evn[TimestampFieldName] = t
			}
		case zerolog.MessageFieldName:
			evn[ShortMessageFieldName] = v
		case zerolog.CallerFieldName:
			if f, l, err := parseCaller(v.(string)); err == nil {
				evn[FileFieldName] = f
				evn[LineNumberFieldName] = l
			}
		case zerolog.ErrorFieldName:
			evn[ErrorFieldName] = v
		case zerolog.ErrorStackFieldName:
			evn[ErrorStackFieldName] = v
		default:
			key, err := formatKey(k)
			if err == ErrorKeyNotAllowed {
				return
			} else if err != nil {
				continue
			}
			if key != k {
				evn[key] = v
			}
		}
	}
	evn[VersionFieldName] = GelfVersion
	evn[HostFieldName] = w.host

	fmt.Printf("time: %f\n", evn[TimestampFieldName])

	w.bufferEvent(evn)
}

func (w *GelfWriter) bufferEvent(evt map[string]interface{}) {
	w.mu.Lock()
	defer w.mu.Unlock()

	d, err := json.Marshal(evt)
	if err != nil {
		fmt.Printf("error marshalling GELF data: %s", err)
		return
	}

	w.buffer.Add(d)
	if w.isBufferSizeExceeded() {
		w.Flush(false)
	}
}

func (w *GelfWriter) isBufferSizeExceeded() bool {
	return w.buffer.size > w.transport.BufferSize()
}

func (w *GelfWriter) sendBuffer(buffer *logBuffer) error {

	err := w.transport.SendBuffer(buffer)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error sending log: %s", err)
		if w.tempLogPath != "" {
			w.writeTemporaryLog(buffer)
		}
	}
	return err
	//_, _ = fmt.Fprintf(os.Stderr, "error sending log: %s", err)
	//addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", w.serverUrl, w.serverPort))
	//if err != nil {
	//    return fmt.Errorf("cannot resolve address: %s", err)
	//}
	//
	//conn, err := net.DialTCP("tcp", nil, addr)
	//if err != nil {
	//    return fmt.Errorf("error dialing: %s", err)
	//}
	//_, err = conn.Write(buffer.Bytes())
	//if err != nil {
	//    return fmt.Errorf("write to server failed: %s", err)
	//}
	//buffer.Reset()
	//return nil
}

func (w *GelfWriter) writeTemporaryLog(buffer *logBuffer) {
	// TODO: implement log save to file
}

func (w *GelfWriter) sendTemporaryLogs() {
	// TODO: send previously saved logs, if any
}

func parseCaller(caller string) (file string, line int, err error) {

	split := strings.Split(caller, ":")
	if len(split) == 2 {
		line, err := strconv.Atoi(split[1])
		return split[0], line, err
	}
	return "", 0, fmt.Errorf("cannot parse caller: %s", caller)
}
