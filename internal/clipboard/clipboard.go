package clipboard

import (
	"bytes"
	"fmt"

	"github.com/atotto/clipboard"
)

// clipboardWriterFunc is an abstraction for the actual clipboard writing function.
type clipboardWriterFunc func(text string) error

// defaultClipboardWriter is the actual function that writes to the system clipboard.
var defaultClipboardWriter clipboardWriterFunc = clipboard.WriteAll

// Writer implements io.WriteCloser. It buffers all data written to it
// until its Close method is called, at which point the buffered content
// is written to the system clipboard using the configured writer function.
type Writer struct {
	buf bytes.Buffer
	// writerFunc is the function used to write to the clipboard.
	// It can be overridden for testing.
	writerFunc clipboardWriterFunc
}

// NewWriter creates and returns a new clipboard Writer.
// The returned Writer buffers all data written to it until Close() is called,
// at which point the buffered content is written to the system clipboard.
func NewWriter() *Writer {
	return &Writer{
		writerFunc: defaultClipboardWriter,
	}
}

// Write appends the given byte slice to the internal buffer.
// It implements the io.Writer interface.
func (cw *Writer) Write(p []byte) (n int, err error) {
	return cw.buf.Write(p)
}

// Close writes all buffered data to the system clipboard using the configured writerFunc.
// It implements the io.Closer interface.
// After writing, the internal buffer is reset.
func (cw *Writer) Close() error {
	if cw.buf.Len() == 0 {
		return nil
	}
	text := cw.buf.String()
	err := cw.writerFunc(text)
	if err != nil {
		return fmt.Errorf("clipboard: failed to write buffered content: %w", err)
	}
	cw.buf.Reset()
	return nil
}
