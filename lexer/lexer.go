package lexer

import (
	"compiler/buffer"
	"io"
	"unicode"
)

// Lexer reads tokens from a source one at a time
type Lexer struct {
	r         *buffer.Lexer
	stack     []ParsingContext
	state     TokenState
	emptyLine bool
}

// NewLexer creates a new Lexer from a io.Reader
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		r:         buffer.NewLexer(r),
		stack:     make([]ParsingContext, 0, 16),
		state:     ExprState,
		emptyLine: true,
	}
}

// Err returns the current error from the buffer reader
func (l *Lexer) Err() error {
	return l.r.Err()
}

func (l *Lexer) enterContext(context ParsingContext) {
	l.stack = append(l.stack, context)
}

func (l *Lexer) leaveContext() ParsingContext {
	ctx := GlobalContext
	if last := len(l.stack) - 1; last >= 0 {
		ctx, l.stack = l.stack[last], l.stack[:last]
	}
	return ctx
}

// Next reads from the buffer and returns the next available token
func (l *Lexer) Next() *Token {
	tt := UnknownToken
	c := l.r.Peek(0)
	switch c {
	case '(':
		if l.state == StmtParensState {
			l.enterContext(StmtParensContext)
		} else {
			l.enterContext(ExprParensContext)
		}
		l.state = ExprState
		l.r.Move(1)
		tt = PunctuatorToken
	case ')':
		if l.leaveContext() == StmtParensContext {
			l.state = ExprState
		} else {
			l.state = SubscriptState
		}
		l.r.Move(1)
		tt = PunctuatorToken
	case '{':
		l.enterContext(BracesContext)
		l.state = ExprState
		l.r.Move(1)
		tt = PunctuatorToken
	case '}':
		if l.leaveContext() == TemplateContext && l.consumeTemplateToken() {
			tt = TemplateToken
		} else {
			l.state = ExprState
			l.r.Move(1)
			tt = PunctuatorToken
		}
	case ']':
		l.state = SubscriptState
		l.r.Move(1)
		tt = PunctuatorToken
	case '[', ';', ',', '~', '?', ':':
		l.state = ExprState
		l.r.Move(1)
		tt = PunctuatorToken
	case '-', '!', '+', '*', '%':
		if l.consumePunctuatorToken() {
			l.state = ExprState
			tt = PunctuatorToken
		}
	case '/':
		if l.consumePunctuatorToken() {
			l.state = ExprState
			tt = PunctuatorToken
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
		if l.consumeNumericToken() {
			tt = NumericToken
			l.state = SubscriptState
		} else if c == '.' {
			l.state = PropNameState
			l.r.Move(1)
			tt = PunctuatorToken
		}
	case ' ', '\t', '\v', '\f':
		l.r.Move(1)
		for l.consumeWhitespace() {
		}
		return &Token{Type: WhitespaceToken, Value: l.r.Shift()}
	case '\n', '\r':
		l.r.Move(1)
		for l.consumeLineTerminator() {
		}
		tt = LineTerminatorToken
	default:
		if l.consumeIdentifierToken() {
			tt = IdentifierToken
		} else if c >= 0xC0 {
			if l.consumeWhitespace() {
				for l.consumeWhitespace() {
				}
				return &Token{Type: WhitespaceToken, Value: l.r.Shift()}
			}
		}
	}

	if tt == UnknownToken {
		_, n := l.r.PeekRune(0)
		l.r.Move(n)
	}

	return &Token{Type: tt, Value: l.r.Shift()}
}

func (l *Lexer) consumePunctuatorToken() bool {
	c := l.r.Peek(0)
	if c == '!' || c == '-' || c == '*' || c == '+' || c == '/' || c == '%' {
		l.r.Move(1)
	}
	return true
}

func (l *Lexer) consumeWhitespace() bool {
	c := l.r.Peek(0)
	l.r.Peek(0)
	if c == ' ' || c == '\t' || c == '\v' || c == '\f' {
		l.r.Move(1)
		return true
	} else if c > 0xC0 {
		if r, n := l.r.PeekRune(0); r == '\u00A0' || r == '\uFFEF' || unicode.Is(unicode.Zs, r) {
			l.r.Move(n)
			return true
		}
	}
	return false
}

func (l *Lexer) consumeLineTerminator() bool {
	c := l.r.Peek(0)
	if c == '\n' {
		l.r.Move(1)
		return true
	} else if c == '\r' {
		if l.r.Peek(1) == '\n' {
			l.r.Move(2)
		} else {
			l.r.Move(1)
		}
		return true
	} else if c > 0xC0 {
		if r, n := l.r.PeekRune(0); r == '\u2028' || r == '\u2029' {
			l.r.Move(n)
			return true
		}
	}
	return false
}

func (l *Lexer) consumeTemplateToken() bool {
	return false
}

func (l *Lexer) consumeNumericToken() bool {
	if l.consumeDigit() {
		for l.consumeDigit() {
		}
		return true
	}
	return false
}

func (l *Lexer) consumeDigit() bool {
	if c := l.r.Peek(0); c >= '0' && c <= '9' {
		l.r.Move(1)
		return true
	}
	return false
}

func (l *Lexer) consumeIdentifierToken() bool {
	c := l.r.Peek(0)
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '$' || c == '_' {
		l.r.Move(1)
	} else if c < 0xC0 {
		return false
	} else {
		return false
	}
	// Deal with unicode
	for {
		c := l.r.Peek(0)
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '$' || c == '_' {
			l.r.Move(1)
		} else {
			break
		}
	}
	return true
}
