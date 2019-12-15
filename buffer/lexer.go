package buffer

import (
	"io"
	"io/ioutil"
)

var nullBuffer = []byte{0}

// Lexer holds the input buffer
type Lexer struct {
	buf     []byte
	pos     int
	start   int
	err     error
	restore func()
}

func NewLexer(r io.Reader) *Lexer {
	var b []byte
	if r != nil {
		if buffer, ok := r.(interface {
			Bytes() []byte
		}); ok {
			b = buffer.Bytes()
		} else {
			var err error
			b, err = ioutil.ReadAll(r)
			if err != nil {
				return &Lexer{
					buf: []byte{0},
					err: err,
				}
			}
		}
	}
	return NewLexerBytes(b)
}

func NewLexerBytes(b []byte) *Lexer {
	z := &Lexer{
		buf: b,
	}

	n := len(b)
	if n == 0 {
		z.buf = nullBuffer
	} else {
		if cap(b) > n {
			b = b[:n+1]
			c := b[n]
			b[n] = 0
			z.buf = b
			z.restore = func() {
				b[n] = c
			}
		} else {
			z.buf = append(b, 0)
		}
	}
	return z
}

// Err returns the error returned from io.Reader or io.EOF when the end has been reached.
func (z *Lexer) Err() error {
	return z.PeekErr(0)
}

func (z *Lexer) Restore() {
	if z.restore != nil {
		z.restore()
		z.restore = nil
	}
}

func (z *Lexer) PeekErr(pos int) error {
	if z.err != nil {
		return z.err
	} else if z.pos+pos >= len(z.buf)-1 {
		return io.EOF
	}
	return nil
}

func (z *Lexer) Peek(pos int) byte {
	pos += z.pos
	return z.buf[pos]
}

func (z *Lexer) PeekRune(pos int) (rune, int) {
	c := z.Peek(pos)
	if c < 0xC0 || z.Peek(pos+1) == 0 {
		return rune(c), 1
	} else if c < 0xE0 || z.Peek(pos+2) == 0 {
		return rune(c&0xF0)<<6 | rune(z.Peek(pos+1)&0x3F), 2
	} else if c < 0xF0 || z.Peek(pos+3) == 0 {
		return rune(c&0xF0)<<12 | rune(z.Peek(pos+1)&0x3F)<<6 | rune(z.Peek(pos+2)&0x3F), 3
	}
	return rune(c&0x07)<<18 | rune(z.Peek(pos+1)&0x3F)<<12 | rune(z.Peek(pos+2)&0x3F)<<6 | rune(z.Peek(pos+3)&0x3F), 4
}

func (z *Lexer) Move(n int) {
	z.pos += n
}

func (z *Lexer) Pos() int {
	return z.pos - z.start
}

func (z *Lexer) Rewind(pos int) {
	z.pos = z.start + pos
}

func (z *Lexer) Lexeme() []byte {
	return z.buf[z.start:z.pos]
}

func (z *Lexer) Skip() {
	z.start = z.pos
}

func (z *Lexer) Shift() []byte {
	b := z.buf[z.start:z.pos]
	z.start = z.pos
	return b
}

func (z *Lexer) Offset() int {
	return z.pos
}

func (z *Lexer) Bytes() []byte {
	return z.buf[:len(z.buf)-1]
}
