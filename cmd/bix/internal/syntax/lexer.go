// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package syntax

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"github.com/billchow98/bixscript/cmd/bix/internal/token"
	"regexp"
	"unicode"
	"unicode/utf8"
)

type lexerErrorHandler func(err error)

type lexer struct {
	source   string
	start    int
	current  int
	file     string
	line     int
	startCol int
	curCol   int
	eh       lexerErrorHandler
}

var numberRe = regexp.MustCompile(`^[+-]?(\d+(\.\d*)?|\.\d+)`)
var newlineRe = regexp.MustCompile(`^\r?\n`)

func newLexer(s string, file string, eh lexerErrorHandler) *lexer {
	l := &lexer{source: s, start: 0, current: 0, file: file, line: 1, startCol: 1, curCol: 1, eh: eh}

	if !utf8.ValidString(s) {
		l.newError("input contains invalid UTF-8 characters")
	}

	return l
}

func (l *lexer) newToken(t token.Type) *token.Token {
	return &token.Token{Type: t, Value: l.source[l.start:l.current], File: l.file, Line: l.line, Col: l.startCol}
}

func (l *lexer) newError(s string) {
	l.eh(newError(s, l.newToken(token.Error)))
}

func (l *lexer) atEnd() bool {
	return l.current == len(l.source)
}

// Returns current (unseen) rune and moves pointer to next new rune
func (l *lexer) nextRune() rune {
	if l.atEnd() {
		// It doesn't matter. If the user really wanted to type this in,
		// it would have been part of a string.
		return '\x00'
	}

	r, size := utf8.DecodeRuneInString(l.source[l.current:])
	if r == utf8.RuneError {
		assert.Assert(size == 1)
		l.newError("error decoding UTF-8 character")
	}

	l.current += size
	l.curCol++

	return r
}

// Returns current (unseen) rune
func (l *lexer) peekRune() rune {
	if l.atEnd() {
		// It doesn't matter. If the user really wanted to type this in,
		// it would have been part of a string.
		return '\x00'
	}

	r, size := utf8.DecodeRuneInString(l.source[l.current:])
	if r == utf8.RuneError {
		assert.Assert(size == 1)
		l.newError("error decoding UTF-8 character")
	}

	return r
}

func (l *lexer) peekNextRune() rune {
	if l.atEnd() {
		// It doesn't matter. If the user really wanted to type this in,
		// it would have been part of a string.
		return '\x00'
	}

	r, size := utf8.DecodeRuneInString(l.source[l.current:])
	if r == utf8.RuneError {
		assert.Assert(size == 1)
		l.newError("error decoding UTF-8 character")
	}

	r, size = utf8.DecodeRuneInString(l.source[l.current+size:])
	if r == utf8.RuneError {
		assert.Assert(size == 1)
		l.newError("error decoding UTF-8 character")
	}

	return r
}

// Advances lexer to next rune
func (l *lexer) advance() {
	_ = l.nextRune()
}

func (l *lexer) nextIsNewline() bool {
	return newlineRe.FindStringIndex(l.source[l.start:]) != nil
}

func (l *lexer) nextNewline() *token.Token {
	length := newlineRe.FindStringIndex(l.source[l.start:])[1]
	l.current = l.start + length
	t := l.newToken(token.Newline)
	l.line++
	l.curCol = 1
	return t
}

// Only for backslash handling
func (l *lexer) skipNewline() {
	_ = l.nextNewline()
}

// Returns a Newline token if following input contains whitespace followed by a newline
// If not, returns nil
func (l *lexer) skipWhitespace() *token.Token {
	inLineComment := false

loop:
	for !l.atEnd() {
		l.start = l.current
		l.startCol = l.curCol

		r := l.peekRune()

		switch {
		case inLineComment && !l.nextIsNewline():
			l.advance()
		case r == '/' && l.peekNextRune() == '/':
			inLineComment = true
			l.advance()
		case !unicode.IsSpace(r):
			break loop
		case l.nextIsNewline():
			return l.nextNewline()
		default:
			l.advance()
		}
	}

	return nil
}

func (l *lexer) nextIsNumber() bool {
	return numberRe.FindStringIndex(l.source[l.start:]) != nil
}

func (l *lexer) nextNumber() *token.Token {
	length := numberRe.FindStringIndex(l.source[l.start:])[1]
	l.current = l.start + length
	return l.newToken(token.Number)
}

func isHex(r rune) bool {
	l := unicode.ToLower(r)
	return '0' <= r && r <= '9' || 'a' <= l && l <= 'f'
}

func (l *lexer) skipHex(n int) {
	for i := 0; i < n; i++ {
		r := l.nextRune()

		if !isHex(r) {
			l.newError("invalid character in hexadecimal escape")
		}
	}
}

func (l *lexer) skipEscape() {
	if l.atEnd() {
		l.newError("unterminated string")
	}

	r := l.nextRune()

	switch r {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '"':
		return
	case 'x':
		l.skipHex(2)
	case 'u':
		l.skipHex(4)
	case 'U':
		l.skipHex(8)
	default:
		l.newError("unknown escape")
	}
}

func (l *lexer) nextString() *token.Token {
	for !l.atEnd() {
		r := l.nextRune()
		switch r {
		case '"':
			return l.newToken(token.String)
		case '\n':
			l.newError("newline in string")
			return l.newToken(token.Error)
		case '\\':
			l.skipEscape()
		}
	}

	l.newError("unterminated string")
	return l.newToken(token.Error)
}

func (l *lexer) nextIdentifier() *token.Token {
	for unicode.IsLetter(l.peekRune()) || unicode.IsDigit(l.peekRune()) {
		l.advance()
	}

	return l.newToken(token.Identifier)
}

// All keywords only use ASCII characters
// By Go's UTF-8 specification, the previous rune uses only 1 byte
func (l *lexer) nextKeyword(t token.Type, s string) *token.Token {
	if l.start+len(s) <= len(l.source) && l.source[l.start:l.start+len(s)] == s {
		l.current = l.start + len(s)
		l.curCol = l.startCol + len(s)
		return l.newToken(t)
	}
	return l.nextIdentifier()
}

// NextToken
// Basically, statements are terminated by either a newline or when EOF is reached.
// Each line should contain at most 1 statement. The semicolon separator does not exist.
// A backslash is always needed to extend a statement to the next line.
func (l *lexer) NextToken() *token.Token {
	l.start = l.current
	l.startCol = l.curCol

	if t := l.skipWhitespace(); t != nil {
		return t
	}

	if l.atEnd() {
		return l.newToken(token.Eof)
	}

	if l.nextIsNumber() {
		return l.nextNumber()
	}

	switch r := l.nextRune(); r {
	case '\\':
		l.start = l.current
		l.startCol = l.curCol

		if l.atEnd() {
			l.newError("unexpected eof after '\\'")
			return l.newToken(token.Error)
		}

		if !l.nextIsNewline() {
			l.newError("stray character after '\\'")
			return l.newToken(token.Error)
		}

		l.skipNewline()

		if t := l.skipWhitespace(); t != nil {
			assert.Assert(t.Type == token.Newline)
			l.line-- // To report correct line number
			l.newError("empty line after '\\'")
			t.Type = token.Error
			l.line++ // To report correct line number
			return t
		}

		return l.NextToken()
	case '+':
		if l.peekRune() == '=' {
			l.advance()
			return l.newToken(token.PlusEqual)
		}
		return l.newToken(token.Plus)
	case '-':
		if l.peekRune() == '=' {
			l.advance()
			return l.newToken(token.MinusEqual)
		}
		return l.newToken(token.Minus)
	case '*':
		if l.peekRune() == '*' {
			l.advance()
			if l.peekRune() == '=' {
				l.advance()
				return l.newToken(token.StarStarEqual)
			}
			return l.newToken(token.StarStar)
		}
		if l.peekRune() == '=' {
			l.advance()
			return l.newToken(token.StarEqual)
		}
		return l.newToken(token.Star)
	case '/':
		if l.peekRune() == '=' {
			l.advance()
			return l.newToken(token.SlashEqual)
		}
		return l.newToken(token.Slash)
	case '!':
		if l.peekRune() == '=' {
			l.advance()
			return l.newToken(token.BangEqual)
		}
		return l.newToken(token.Bang)
	case '=':
		if l.peekRune() == '=' {
			l.advance()
			return l.newToken(token.EqualEqual)
		}
		return l.newToken(token.Equal)
	case '<':
		if l.peekRune() == '=' {
			l.advance()
			return l.newToken(token.LessEqual)
		}
		return l.newToken(token.Less)
	case '>':
		if l.peekRune() == '=' {
			l.advance()
			return l.newToken(token.GreaterEqual)
		}
		return l.newToken(token.Greater)
	case '"':
		return l.nextString()
	case '.':
		return l.newToken(token.Dot)
	case '(':
		return l.newToken(token.LeftParen)
	case ')':
		return l.newToken(token.RightParen)
	case '{':
		return l.newToken(token.LeftBrace)
	case '}':
		return l.newToken(token.RightBrace)
	case ';':
		return l.newToken(token.Semicolon)
	case ',':
		return l.newToken(token.Comma)
	case 'a':
		return l.nextKeyword(token.And, "and")
	case 'e':
		return l.nextKeyword(token.Else, "else")
	case 'f':
		switch l.peekRune() {
		case 'a':
			return l.nextKeyword(token.False, "false")
		case 'o':
			return l.nextKeyword(token.For, "for")
		case 'u':
			return l.nextKeyword(token.Function, "function")
		default:
			return l.nextIdentifier()
		}
	case 'i':
		return l.nextKeyword(token.If, "if")
	case 'l':
		return l.nextKeyword(token.Let, "let")
	case 'o':
		return l.nextKeyword(token.Or, "or")
	case 'r':
		return l.nextKeyword(token.Return, "return")
	case 't':
		return l.nextKeyword(token.True, "true")
	case 'w':
		return l.nextKeyword(token.While, "while")
	default:
		if unicode.IsLetter(r) {
			return l.nextIdentifier()
		}

		l.newError("unknown token")
		return l.newToken(token.Error)
	}
}
