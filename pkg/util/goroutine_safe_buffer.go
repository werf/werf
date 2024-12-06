package util

import (
	"bytes"
	"io"
	"sync"
)

type GoroutineSafeBuffer struct {
	*bytes.Buffer
	m sync.Mutex
}

func NewGoroutineSafeBuffer() *GoroutineSafeBuffer {
	return &GoroutineSafeBuffer{Buffer: bytes.NewBuffer([]byte{})}
}

func (b *GoroutineSafeBuffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.Read(p)
}

func (b *GoroutineSafeBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.Write(p)
}

func (b *GoroutineSafeBuffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.String()
}

func (b *GoroutineSafeBuffer) Bytes() []byte {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.Bytes()
}

func (b *GoroutineSafeBuffer) Cap() int {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.Cap()
}

func (b *GoroutineSafeBuffer) Grow(n int) {
	b.m.Lock()
	defer b.m.Unlock()
	b.Buffer.Grow(n)
}

func (b *GoroutineSafeBuffer) Len() int {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.Len()
}

func (b *GoroutineSafeBuffer) Next(n int) []byte {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.Next(n)
}

func (b *GoroutineSafeBuffer) ReadByte() (c byte, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.ReadByte()
}

func (b *GoroutineSafeBuffer) ReadBytes(delim byte) (line []byte, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.ReadBytes(delim)
}

func (b *GoroutineSafeBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.ReadFrom(r)
}

func (b *GoroutineSafeBuffer) ReadRune() (r rune, size int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.ReadRune()
}

func (b *GoroutineSafeBuffer) ReadString(delim byte) (line string, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.ReadString(delim)
}

func (b *GoroutineSafeBuffer) Reset() {
	b.m.Lock()
	defer b.m.Unlock()
	b.Buffer.Reset()
}

func (b *GoroutineSafeBuffer) Truncate(n int) {
	b.m.Lock()
	defer b.m.Unlock()
	b.Buffer.Truncate(n)
}

func (b *GoroutineSafeBuffer) UnreadByte() error {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.UnreadByte()
}

func (b *GoroutineSafeBuffer) UnreadRune() error {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.UnreadRune()
}

func (b *GoroutineSafeBuffer) WriteByte(c byte) error {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.WriteByte(c)
}

func (b *GoroutineSafeBuffer) WriteRune(r rune) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.WriteRune(r)
}

func (b *GoroutineSafeBuffer) WriteString(s string) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.WriteString(s)
}

func (b *GoroutineSafeBuffer) WriteTo(w io.Writer) (n int64, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.Buffer.WriteTo(w)
}
