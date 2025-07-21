package analyzer

import (
	"fmt"
	"unicode"
)

// TokenType represents the type of a token
type TokenType int

// TokenType constants define the types of tokens in effect expressions
const (
	TokenEOF     TokenType = iota
	TokenLBrace      // {
	TokenRBrace      // }
	TokenLParen      // (
	TokenRParen      // )
	TokenPipe        // |
	TokenLBracket    // [
	TokenRBracket    // ]
	TokenIdent       // identifier
	TokenIllegal     // illegal token
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
	return l.input[start : l.pos-1]
}

// NextToken returns the next token
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	tok.Pos = l.pos - 1

	switch l.ch {
	case '{':
		tok.Type = TokenLBrace
		tok.Value = "{"
		l.readChar()
	case '}':
		tok.Type = TokenRBrace
		tok.Value = "}"
		l.readChar()
	case '(':
		tok.Type = TokenLParen
		tok.Value = "("
		l.readChar()
	case ')':
		tok.Type = TokenRParen
		tok.Value = ")"
		l.readChar()
	case '|':
		tok.Type = TokenPipe
		tok.Value = "|"
		l.readChar()
	case '[':
		tok.Type = TokenLBracket
		tok.Value = "["
		l.readChar()
	case ']':
		tok.Type = TokenRBracket
		tok.Value = "]"
		l.readChar()
	case 0:
		tok.Type = TokenEOF
		tok.Value = ""
	default:
		if isLetter(l.ch) {
			tok.Value = l.readIdentifier()
			tok.Type = TokenIdent
			return tok
		}
		tok.Type = TokenIllegal
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
	case TokenEOF:
		return "EOF"
	case TokenLBrace:
		return "{"
	case TokenRBrace:
		return "}"
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenPipe:
		return "|"
	case TokenLBracket:
		return "["
	case TokenRBracket:
		return "]"
	case TokenIdent:
		return "IDENT"
	case TokenIllegal:
		return "ILLEGAL"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

// String returns a string representation of a token
func (t Token) String() string {
	if t.Type == TokenIdent || t.Type == TokenIllegal {
		return fmt.Sprintf("%s(%s)", TokenString(t.Type), t.Value)
	}
	return TokenString(t.Type)
}
