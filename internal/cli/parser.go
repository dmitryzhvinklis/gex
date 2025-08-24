package cli

import (
	"errors"
	"strings"
	"unicode"
)

// Command represents a parsed command with arguments
type Command struct {
	Name       string
	Args       []string
	Pipes      []*Command
	Redirect   *Redirect
	Background bool
}

// Redirect represents input/output redirection
type Redirect struct {
	Type   RedirectType
	Target string
}

type RedirectType int

const (
	RedirectNone   RedirectType = iota
	RedirectOut                 // >
	RedirectAppend              // >>
	RedirectIn                  // <
	RedirectErr                 // 2>
	RedirectBoth                // &>
)

// Parser provides high-performance command parsing
type Parser struct {
	input  string
	pos    int
	length int
}

// Parse parses a command line input into a Command structure
func Parse(input string) (*Command, error) {
	if input == "" {
		return nil, errors.New("empty command")
	}

	p := &Parser{
		input:  input,
		pos:    0,
		length: len(input),
	}

	return p.parseCommand()
}

// parseCommand parses the main command and handles pipes
func (p *Parser) parseCommand() (*Command, error) {
	cmd, err := p.parseSimpleCommand()
	if err != nil {
		return nil, err
	}

	// Handle pipes
	for p.pos < p.length && p.peek() == '|' {
		p.advance() // consume '|'
		p.skipWhitespace()

		nextCmd, err := p.parseSimpleCommand()
		if err != nil {
			return nil, err
		}

		cmd.Pipes = append(cmd.Pipes, nextCmd)
	}

	return cmd, nil
}

// parseSimpleCommand parses a single command without pipes
func (p *Parser) parseSimpleCommand() (*Command, error) {
	p.skipWhitespace()

	if p.pos >= p.length {
		return nil, errors.New("unexpected end of input")
	}

	cmd := &Command{}

	// Parse command name
	name, err := p.parseToken()
	if err != nil {
		return nil, err
	}
	cmd.Name = name

	// Parse arguments and redirections
	for p.pos < p.length {
		p.skipWhitespace()

		if p.pos >= p.length {
			break
		}

		ch := p.peek()

		// Handle background execution
		if ch == '&' && p.pos == p.length-1 {
			cmd.Background = true
			p.advance()
			break
		}

		// Handle pipes - return to parent
		if ch == '|' {
			break
		}

		// Handle redirections
		if redirect := p.parseRedirect(); redirect != nil {
			cmd.Redirect = redirect
			continue
		}

		// Parse argument
		arg, err := p.parseToken()
		if err != nil {
			return nil, err
		}
		cmd.Args = append(cmd.Args, arg)
	}

	return cmd, nil
}

// parseToken parses a single token (command name or argument)
func (p *Parser) parseToken() (string, error) {
	p.skipWhitespace()

	if p.pos >= p.length {
		return "", errors.New("unexpected end of input")
	}

	var result strings.Builder
	quoted := false
	quoteChar := byte(0)

	for p.pos < p.length {
		ch := p.current()

		// Handle quotes
		if !quoted && (ch == '"' || ch == '\'') {
			quoted = true
			quoteChar = ch
			p.advance()
			continue
		}

		if quoted && ch == quoteChar {
			quoted = false
			quoteChar = 0
			p.advance()
			continue
		}

		// Handle escape sequences
		if ch == '\\' && p.pos+1 < p.length {
			p.advance()
			next := p.current()
			result.WriteByte(next)
			p.advance()
			continue
		}

		// Break on whitespace or special characters if not quoted
		if !quoted {
			if unicode.IsSpace(rune(ch)) || ch == '|' || ch == '>' || ch == '<' || ch == '&' {
				break
			}
		}

		result.WriteByte(ch)
		p.advance()
	}

	if quoted {
		return "", errors.New("unterminated quote")
	}

	token := result.String()
	if token == "" {
		return "", errors.New("empty token")
	}

	return token, nil
}

// parseRedirect parses redirection operators
func (p *Parser) parseRedirect() *Redirect {
	if p.pos >= p.length {
		return nil
	}

	ch := p.current()

	switch ch {
	case '>':
		p.advance()
		if p.pos < p.length && p.current() == '>' {
			p.advance()
			return &Redirect{Type: RedirectAppend, Target: p.parseRedirectTarget()}
		}
		return &Redirect{Type: RedirectOut, Target: p.parseRedirectTarget()}

	case '<':
		p.advance()
		return &Redirect{Type: RedirectIn, Target: p.parseRedirectTarget()}

	case '2':
		if p.pos+1 < p.length && p.input[p.pos+1] == '>' {
			p.advance() // consume '2'
			p.advance() // consume '>'
			return &Redirect{Type: RedirectErr, Target: p.parseRedirectTarget()}
		}

	case '&':
		if p.pos+1 < p.length && p.input[p.pos+1] == '>' {
			p.advance() // consume '&'
			p.advance() // consume '>'
			return &Redirect{Type: RedirectBoth, Target: p.parseRedirectTarget()}
		}
	}

	return nil
}

// parseRedirectTarget parses the target of a redirection
func (p *Parser) parseRedirectTarget() string {
	p.skipWhitespace()

	if p.pos >= p.length {
		return ""
	}

	var result strings.Builder

	for p.pos < p.length {
		ch := p.current()

		if unicode.IsSpace(rune(ch)) || ch == '|' || ch == '&' {
			break
		}

		result.WriteByte(ch)
		p.advance()
	}

	return result.String()
}

// Helper methods for efficient parsing
func (p *Parser) current() byte {
	if p.pos >= p.length {
		return 0
	}
	return p.input[p.pos]
}

func (p *Parser) peek() byte {
	return p.current()
}

func (p *Parser) advance() {
	if p.pos < p.length {
		p.pos++
	}
}

func (p *Parser) skipWhitespace() {
	for p.pos < p.length && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}
