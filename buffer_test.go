package zgelf

import (
	"reflect"
	"testing"
)

func TestNewLogBuffer(t *testing.T) {
	tests := []struct {
		name string
		want *logBuffer
	}{
		{"create new logBuffer", &logBuffer{
			buffers:   make([][]byte, 0),
			size:      0,
			compress:  false,
			delimiter: make([]byte, 0),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLogBuffer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLogBuffer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_logBuffer_Add(t *testing.T) {
	tests := []struct {
		name         string
		buffer       *logBuffer
		data         []byte
		expectedSize int
		expectedData []byte
	}{
		{"add to empty buffer", NewLogBuffer(), []byte(`{ "x": "Hello World" }`),
			22, []byte(`{ "x": "Hello World" }`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.buffer.Add(tt.data)

			if tt.buffer.size != tt.expectedSize {
				t.Errorf("buffer.size error, want %d got: %d",
					tt.expectedSize, tt.buffer.size)
			}

		})
	}
}

func Test_logBuffer_Clear(t *testing.T) {

	b1 := NewLogBuffer()
	b2 := NewLogBuffer()

	b1.Add([]byte(`{ "x":"Hello World" }`))
	b1.compress = true
	b2.Add([]byte(`{ "x":"Hello World" }`))
	b2.SetDelimiter([]byte{'\n'})

	tests := []struct {
		name   string
		buffer *logBuffer
	}{
		{"clear 1", b1},
		{"clear 2", b2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.buffer.Clear()
			if tt.buffer.size > 0 {
				t.Errorf("buffer.size error, want 0 got: %d", tt.buffer.size)
			}
			if tt.buffer.compress {
				t.Errorf("buffer.compress error, want false got: %t", tt.buffer.compress)
			}
			if len(tt.buffer.buffers) > 0 {
				t.Errorf("buffer.buffers not empty, got: %v", tt.buffer.buffers)
			}
			if len(tt.buffer.delimiter) > 0 {
				t.Errorf("buffer.delimiter not empty, got: %v", tt.buffer.delimiter)
			}
		})
	}
}

func Test_logBuffer_Copy(t *testing.T) {
	//type fields struct {
	//    buffers   [][]byte
	//    size      int
	//    compress  bool
	//    delimiter []byte
	//}
	//tests := []struct {
	//    name   string
	//    fields fields
	//    want   *logBuffer
	//}{
	//    // TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//    t.Run(tt.name, func(t *testing.T) {
	//        b := &logBuffer{
	//            buffers:   tt.fields.buffers,
	//            size:      tt.fields.size,
	//            compress:  tt.fields.compress,
	//            delimiter: tt.fields.delimiter,
	//        }
	//        if got := b.Copy(); !reflect.DeepEqual(got, tt.want) {
	//            t.Errorf("Copy() = %v, want %v", got, tt.want)
	//        }
	//    })
	//}
}

func Test_logBuffer_EnableCompression(t *testing.T) {
	//type fields struct {
	//    buffers   [][]byte
	//    size      int
	//    compress  bool
	//    delimiter []byte
	//}
	//tests := []struct {
	//    name   string
	//    fields fields
	//}{
	//    // TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//    t.Run(tt.name, func(t *testing.T) {
	//        b := &logBuffer{
	//            buffers:   tt.fields.buffers,
	//            size:      tt.fields.size,
	//            compress:  tt.fields.compress,
	//            delimiter: tt.fields.delimiter,
	//        }
	//    })
	//}
}

func Test_logBuffer_Read(t *testing.T) {
	//type fields struct {
	//    buffers   [][]byte
	//    size      int
	//    compress  bool
	//    delimiter []byte
	//}
	//type args struct {
	//    p []byte
	//}
	//tests := []struct {
	//    name    string
	//    fields  fields
	//    args    args
	//    wantN   int
	//    wantErr bool
	//}{
	//    // TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//    t.Run(tt.name, func(t *testing.T) {
	//        b := &logBuffer{
	//            buffers:   tt.fields.buffers,
	//            size:      tt.fields.size,
	//            compress:  tt.fields.compress,
	//            delimiter: tt.fields.delimiter,
	//        }
	//        gotN, err := b.Read(tt.args.p)
	//        if (err != nil) != tt.wantErr {
	//            t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
	//            return
	//        }
	//        if gotN != tt.wantN {
	//            t.Errorf("Read() gotN = %v, want %v", gotN, tt.wantN)
	//        }
	//    })
	//}
}

func Test_logBuffer_SetDelimiter(t *testing.T) {
	//type fields struct {
	//    buffers   [][]byte
	//    size      int
	//    compress  bool
	//    delimiter []byte
	//}
	//type args struct {
	//    d []byte
	//}
	//tests := []struct {
	//    name   string
	//    fields fields
	//    args   args
	//}{
	//    // TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//    t.Run(tt.name, func(t *testing.T) {
	//        b := &logBuffer{
	//            buffers:   tt.fields.buffers,
	//            size:      tt.fields.size,
	//            compress:  tt.fields.compress,
	//            delimiter: tt.fields.delimiter,
	//        }
	//    })
	//}
}

func Test_logBuffer_Size(t *testing.T) {
	//type fields struct {
	//    buffers   [][]byte
	//    size      int
	//    compress  bool
	//    delimiter []byte
	//}
	//tests := []struct {
	//    name   string
	//    fields fields
	//    want   int
	//}{
	//    // TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//    t.Run(tt.name, func(t *testing.T) {
	//        b := &logBuffer{
	//            buffers:   tt.fields.buffers,
	//            size:      tt.fields.size,
	//            compress:  tt.fields.compress,
	//            delimiter: tt.fields.delimiter,
	//        }
	//        if got := b.Size(); got != tt.want {
	//            t.Errorf("Size() = %v, want %v", got, tt.want)
	//        }
	//    })
	//}
}
