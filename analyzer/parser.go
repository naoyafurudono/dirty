package analyzer

import (
	"fmt"
	"strings"
)

// Parser parses effect declarations
type Parser struct {
	lexer *Lexer
	cur   Token
	peek  Token
}

// NewParser creates a new parser
func NewParser(input string) *Parser {
	p := &Parser{
		lexer: NewLexer(input),
	}
	// Read two tokens to initialize cur and peek
	p.nextToken()
	p.nextToken()
	return p
}

// nextToken advances to the next token
func (p *Parser) nextToken() {
	p.cur = p.peek
	p.peek = p.lexer.NextToken()
}

// ParseEffectDecl parses an effect declaration
// The input should be the content after "//dirty:" or "// dirty:"
func ParseEffectDecl(comment string) (EffectExpr, error) {
	// Remove "//dirty:" or "// dirty:" prefix and trim spaces
	comment = strings.TrimSpace(comment)
	content := strings.TrimPrefix(comment, "//dirty:")
	if content == comment {
		// Try with space
		content = strings.TrimPrefix(comment, "// dirty:")
	}
	content = strings.TrimSpace(content)
	
	if content == "" {
		// Empty declaration means empty set
		return &LiteralSet{Elements: []EffectExpr{}}, nil
	}
	
	parser := NewParser(content)
	return parser.parseSetExpr()
}

// parseSetExpr parses a set expression: { ... }
func (p *Parser) parseSetExpr() (EffectExpr, error) {
	if p.cur.Type != TOKEN_LBRACE {
		return nil, fmt.Errorf("expected '{' at position %d, got %s", p.cur.Pos, p.cur.String())
	}
	p.nextToken() // skip {
	
	// Handle empty set
	if p.cur.Type == TOKEN_RBRACE {
		p.nextToken() // skip }
		return &LiteralSet{Elements: []EffectExpr{}}, nil
	}
	
	// Parse elements
	elements := []EffectExpr{}
	for {
		elem, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
		
		// Check for | or }
		if p.cur.Type == TOKEN_PIPE {
			p.nextToken() // skip |
			continue
		}
		
		if p.cur.Type == TOKEN_RBRACE {
			p.nextToken() // skip }
			break
		}
		
		return nil, fmt.Errorf("expected '|' or '}' at position %d, got %s", p.cur.Pos, p.cur.String())
	}
	
	return &LiteralSet{Elements: elements}, nil
}

// parsePrimary parses a primary expression
func (p *Parser) parsePrimary() (EffectExpr, error) {
	switch p.cur.Type {
	case TOKEN_IDENT:
		// Could be an effect label or effect reference
		ident := p.cur.Value
		p.nextToken()
		
		if p.cur.Type == TOKEN_LBRACKET {
			// Effect label: operation[target]
			p.nextToken() // skip [
			
			if p.cur.Type != TOKEN_IDENT {
				return nil, fmt.Errorf("expected identifier after '[' at position %d, got %s", p.cur.Pos, p.cur.String())
			}
			target := p.cur.Value
			p.nextToken()
			
			if p.cur.Type != TOKEN_RBRACKET {
				return nil, fmt.Errorf("expected ']' at position %d, got %s", p.cur.Pos, p.cur.String())
			}
			p.nextToken() // skip ]
			
			return &EffectLabel{
				Operation: ident,
				Target:    target,
			}, nil
		}
		
		// Effect reference (Phase 2, but parse it anyway)
		return &EffectRef{Name: ident}, nil
		
	case TOKEN_LPAREN:
		// Parenthesized expression
		p.nextToken() // skip (
		expr, err := p.parseUnionExpr()
		if err != nil {
			return nil, err
		}
		if p.cur.Type != TOKEN_RPAREN {
			return nil, fmt.Errorf("expected ')' at position %d, got %s", p.cur.Pos, p.cur.String())
		}
		p.nextToken() // skip )
		return expr, nil
		
	default:
		return nil, fmt.Errorf("unexpected token at position %d: %s", p.cur.Pos, p.cur.String())
	}
}

// parseUnionExpr parses a union expression (for future use with parentheses)
func (p *Parser) parseUnionExpr() (EffectExpr, error) {
	elements := []EffectExpr{}
	
	for {
		elem, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
		
		if p.cur.Type != TOKEN_PIPE {
			break
		}
		p.nextToken() // skip |
	}
	
	// If only one element, return it directly
	if len(elements) == 1 {
		return elements[0], nil
	}
	
	// Otherwise, wrap in a literal set
	return &LiteralSet{Elements: elements}, nil
}