package zgelf

import (
    "encoding/json"
    "fmt"
    "github.com/rs/zerolog"
    "strings"
    "unicode"
)

func convertTime(i json.Number, timeFormat string) (float64, error) {
    if timeFormat == zerolog.TimeFormatUnix {
        return i.Float64()
    } else if timeFormat == zerolog.TimeFormatUnixMs {
        if f, err := i.Float64(); err != nil {
            return 0, err
        } else {
            return f / 1000.0, nil
        }
    } else if timeFormat == zerolog.TimeFormatUnixMicro {
        if f, err := i.Int64(); err != nil {
            return 0, err
        } else {
            return float64(f/1000) / 1000.0, nil
        }
    }
    return 0, fmt.Errorf("unknown timeformat")
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

// parseLogLevel converts the zeroLog-level to a syslog level
//   0    Emergency
//   1    Alert
//   2    Critical
//   3    Error
//   4    Warning
//   5    Notice
//   6    Informational
//   7    Debug
func parseLogLevel(level string) int {
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
