package zgelf

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"testing"
)

func Test_convertTime(t *testing.T) {
	type args struct {
		i          json.Number
		timeFormat string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{"Format Unix", args{json.Number("1234567890"), zerolog.TimeFormatUnix}, float64(1234567890), false},
		{"Format UnixMs", args{json.Number("1234567890"), zerolog.TimeFormatUnixMs}, 1234567.890, false},
		{"Format UnixMicro", args{json.Number("1234567890"), zerolog.TimeFormatUnixMicro}, 1234.567, false},
		{"Format UnknownFormat", args{json.Number("1234567890"), "foobar"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertTime(tt.args.i, tt.args.timeFormat)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("convertTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formatKey(t *testing.T) {
	type args struct {
		k string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Format lowercase short key", args{"key"}, "_key", false},
		{"Format lowercase long key", args{"key_long"}, "_key_long", false},
		{"Format camelCase long key", args{"keyCamel"}, "_key_camel", false},
		{"Format PascalCase long key", args{"KeyPascal"}, "_key_pascal", false},
		{"Format key 1", args{"key.field"}, "_key.field", false},
		{"Format key 2", args{"-key"}, "_-key", false},
		{"Format key 3", args{"key.fIEld"}, "_key.f_i_eld", false},
		{"Format not allowed key 1", args{"Id"}, "", true},
		{"Format not allowed key 2", args{"id"}, "", true},
		{"Format not allowed key 3", args{"_id"}, "", true},
		{"Format empty key", args{""}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatKey(tt.args.k)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("formatKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseLogLevel(t *testing.T) {
	type args struct {
		level string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"parse debug level", args{"debug"}, 7},
		{"parse info level", args{"info"}, 6},
		{"parse warn level", args{"warn"}, 4},
		{"parse error level", args{"error"}, 3},
		{"parse fatal level", args{"fatal"}, 2},
		{"parse panic level", args{"panic"}, 1},
		{"parse invalid level", args{"invalid"}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseLogLevel(tt.args.level); got != tt.want {
				t.Errorf("parseLogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}
