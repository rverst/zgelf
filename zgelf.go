package zgelf

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/rs/zerolog"
    "net"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
    "unicode"
)

const (
    tempLogFileRegex       = "^log_([[:digit:]]+).gz$"
    defaultBufferSize      = 1024 * 1024     // default buffer size  (flush if size exceeds 1MB)
    defaultBufferTime      = time.Minute * 5 // default buffer time (flush every 5 minutes)
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
    TempLogPath string

    bufferSize int
    bufferTime time.Duration
    serverUrl  string
    serverPort uint16
    host       string
    queue      chan map[string]interface{}
    wg         sync.WaitGroup
    mu         sync.RWMutex
    buffer     *bytes.Buffer
    ticker     *time.Ticker
}

func (w *GelfWriter) isEmpty() bool {
    return w.buffer.Len() == 0
}

func NewGelfWriter(host, url string, port uint16) GelfWriter {

    w := GelfWriter{
        TempLogPath: "",
        bufferSize:  defaultBufferSize,
        bufferTime:  defaultBufferTime,
        serverUrl:   url,
        serverPort:  port,
        host:        host,
        buffer:      new(bytes.Buffer),
        ticker:      time.NewTicker(defaultBufferTime),
        queue:       make(chan map[string]interface{}, 500),
    }

    go func() {
        for range w.ticker.C {
            w.Flush()
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
    if l > 0 && l % 50 == 0 {
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
    w.Flush()
    if !w.isEmpty() {
        fmt.Println("BUFFER NOT EMPTY")
    }
}

// Flush flushes the send buffer
func (w *GelfWriter) Flush() {

    time.Sleep(time.Millisecond * 5)
    w.mu.Lock()
    defer w.mu.Unlock()

    if w.isEmpty() {
        return
    }

    err := w.sendLog(w.buffer)
    if err != nil {
        _, _ = fmt.Fprintf(os.Stderr, "error sending log: %s", err)
        if w.TempLogPath != "" {
            w.writeTemporaryLog(w.buffer)
        }
    } else {
        w.sendTemporaryLogs()
    }

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
            lvl := parseLevel(v.(string))
            if lvl < 0 {
                return
            }
            evn[LevelFieldName] = lvl
            evn[OriginalLevelFieldName] = v
        case zerolog.TimestampFieldName:
            t, err := convertTime(v.(json.Number))
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

func (w *GelfWriter) SetMaxBufferTime(bufferTime time.Duration) {
    if w.ticker != nil {
        w.ticker.Stop()
    }
    w.ticker = time.NewTicker(bufferTime)
}

func (w *GelfWriter) bufferEvent(evt map[string]interface{}) {
    w.mu.Lock()
    defer w.mu.Unlock()

    d, err := json.Marshal(evt)
    if err != nil {
        fmt.Printf("error marshalling GELF data: %s", err)
        return
    }

    if !w.isEmpty() {
        w.buffer.WriteByte(0)
    }

    w.buffer.Write(d)
    if w.isBufferSizeExceeded() {
        w.Flush()
    }

}

func (w *GelfWriter) isBufferSizeExceeded() bool {
    return w.buffer.Len() > w.bufferSize
}

func (w *GelfWriter) sendLog(buffer *bytes.Buffer) error {

    addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", w.serverUrl, w.serverPort))
    if err != nil {
        return fmt.Errorf("cannot resolve address: %s", err)
    }

    conn, err := net.DialTCP("tcp", nil, addr)
    if err != nil {
        return fmt.Errorf("error dialing: %s", err)
    }
    _, err = conn.Write(buffer.Bytes())
    if err != nil {
        return fmt.Errorf("write to server failed: %s", err)
    }
    buffer.Reset()
    return nil
}

func (w *GelfWriter) writeTemporaryLog(buffer *bytes.Buffer) {
    // TODO: implement log save to file
    buffer.Reset()
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

func formatKey(k string) (string, error) {

    var key strings.Builder
    key.Grow(len(k))
    for i, c := range k {
        if i == 0 && c != '_' {
            key.WriteRune('_')
        }
        if unicode.IsLetter(c) || unicode.IsNumber(c) ||
            c == '_' || c == '-' || c == '.' {
            if unicode.IsUpper(c) {
                if i > 0 {
                    key.WriteRune('_')
                }
                key.WriteRune(unicode.ToLower(c))
            } else {
                key.WriteRune(c)
            }
        }
    }

    if key.Len() == 0 {
        return "", fmt.Errorf("cannot convert to valid key: %s", k)
    }

    if key.String() == NotAllowedIdFieldName {
        return "", ErrorKeyNotAllowed
    }
    return key.String(), nil
}

func convertTime(i json.Number) (float64, error) {
    if zerolog.TimeFieldFormat == zerolog.TimeFormatUnix {
        return i.Float64()
    } else if zerolog.TimeFieldFormat == zerolog.TimeFormatUnixMs {
        if f, err := i.Float64(); err != nil {
            return 0, err
        } else {
            return f / 1000.0, nil
        }
    } else if zerolog.TimeFieldFormat == zerolog.TimeFormatUnixMicro {
        if f, err := i.Int64(); err != nil {
            return 0, err
        } else {
            return float64(f / 1000) / 1000.0, nil
        }
    }
    return 0, fmt.Errorf("unknown timeformat")
}

// parseLevel converts the zeroLog-level to a syslog level
//   0    Emergency
//   1    Alert
//   2    Critical
//   3    Error
//   4    Warning
//   5    Notice
//   6    Informational
//   7    Debug
func parseLevel(level string) int {
    lvl, err := zerolog.ParseLevel(level)
    if err != nil {
        return -1
    }
    switch lvl {
    case zerolog.DebugLevel:
        return 7
    case zerolog.InfoLevel:
        return 6
    case zerolog.WarnLevel:
        return 4
    case zerolog.ErrorLevel:
        return 3
    case zerolog.FatalLevel:
        return 2
    case zerolog.PanicLevel:
        return 1
    default:
        return -1
    }
}
