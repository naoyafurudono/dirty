package analyzer

import (
	"fmt"
	"unicode"
)

// TokenType represents the type of a token
type TokenType int

const (
	TOKEN_EOF TokenType = iota
	TOKEN_LBRACE      // {
	TOKEN_RBRACE      // }
	TOKEN_LPAREN      // (
	TOKEN_RPAREN      // )
	TOKEN_PIPE        // |
	TOKEN_LBRACKET    // [
	TOKEN_RBRACKET    // ]
	TOKEN_IDENT       // identifier
	TOKEN_ILLEGAL     // illegal token
)

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
	Pos   int // position in input
}

// Lexer performs lexical analysis
type Lexer struct {
	input string
	pos   int  // current position in input
	ch    rune // current character
}

// NewLexer creates a new lexer for the given input
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// readChar reads the next character
func (l *Lexer) readChar() {
	if l.pos >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = rune(l.input[l.pos])
	}
	l.pos++
}

// peekChar returns the next character without advancing
func (l *Lexer) peekChar() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readIdentifier reads an identifier
func (l *Lexer) readIdentifier() string {
	start := l.pos - 1
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '-' || l.ch == '.' {
		l.readChar()
	}
	return l.input[start:l.pos-1]
}

// NextToken returns the next token
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	tok.Pos = l.pos - 1

	switch l.ch {
	case '{':
		tok.Type = TOKEN_LBRACE
		tok.Value = "{"
		l.readChar()
	case '}':
		tok.Type = TOKEN_RBRACE
		tok.Value = "}"
		l.readChar()
	case '(':
		tok.Type = TOKEN_LPAREN
		tok.Value = "("
		l.readChar()
	case ')':
		tok.Type = TOKEN_RPAREN
		tok.Value = ")"
		l.readChar()
	case '|':
		tok.Type = TOKEN_PIPE
		tok.Value = "|"
		l.readChar()
	case '[':
		tok.Type = TOKEN_LBRACKET
		tok.Value = "["
		l.readChar()
	case ']':
		tok.Type = TOKEN_RBRACKET
		tok.Value = "]"
		l.readChar()
	case 0:
		tok.Type = TOKEN_EOF
		tok.Value = ""
	default:
		if isLetter(l.ch) {
			tok.Value = l.readIdentifier()
			tok.Type = TOKEN_IDENT
			return tok
		}
		tok.Type = TOKEN_ILLEGAL
		tok.Value = string(l.ch)
		l.readChar()
	}

	return tok
}

// isLetter returns true if the rune is a letter
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

// isDigit returns true if the rune is a digit
func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

// TokenString returns a string representation of a token type
func TokenString(t TokenType) string {
	switch t {
	case TOKEN_EOF:
		return "EOF"
	case TOKEN_LBRACE:
		return "{"
	case TOKEN_RBRACE:
		return "}"
	case TOKEN_LPAREN:
		return "("
	case TOKEN_RPAREN:
		return ")"
	case TOKEN_PIPE:
		return "|"
	case TOKEN_LBRACKET:
		return "["
	case TOKEN_RBRACKET:
		return "]"
	case TOKEN_IDENT:
		return "IDENT"
	case TOKEN_ILLEGAL:
		return "ILLEGAL"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

// String returns a string representation of a token
func (t Token) String() string {
	if t.Type == TOKEN_IDENT || t.Type == TOKEN_ILLEGAL {
		return fmt.Sprintf("%s(%s)", TokenString(t.Type), t.Value)
	}
	return TokenString(t.Type)
}
