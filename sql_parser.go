package db_in_go

import "strings"

type Parser struct {
	buf string
	pos int
}

func NewParser(s string) Parser {
	return Parser{buf: strings.TrimSpace(s), pos: 0}
}

func (p *Parser) advance() (ch byte, ok bool) {
	if p.pos >= len(p.buf) {
		return
	}

	p.pos += 1
	if p.pos == len(p.buf) {
		return
	}

	ok = true
	ch = p.current()
	return
}

func (p *Parser) isEnd() bool {
	return p.pos >= len(p.buf)
}

func (p *Parser) current() byte {
	return p.buf[p.pos]
}

func (p *Parser) skipSpace() {
	for !p.isEnd() && isSpace(p.current()) {
		p.advance()
	}
}

func (p *Parser) tryName() (string, bool) {
	if p.isEnd() || !isNameStart(p.current()) {
		return "", false
	}

	start := p.pos
	end := p.pos + 1
	c, ok := p.advance()
	for ok && isNameContinue(c) {
		end += 1
		c, ok = p.advance()
	}

	p.skipSpace()
	return p.buf[start:end], true
}

func (p *Parser) tryKeyword(kw string) bool {
	if len(kw)+p.pos > len(p.buf) ||
		(len(kw)+p.pos != len(p.buf) && !isSeparator(p.buf[p.pos+len(kw)])) {
		return false
	}

	kw = strings.ToLower(kw)
	sub_str := strings.ToLower(p.buf[p.pos : p.pos+len(kw)])
	if kw != sub_str {
		return false
	}

	p.pos += len(kw)
	p.skipSpace()
	return true
}

func isSeparator(ch byte) bool {
	return ch < 128 && !isNameContinue(ch)
}

func isSpace(ch byte) bool {
	switch ch {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

func isAlpha(ch byte) bool {
	return 'a' <= (ch|32) && (ch|32) <= 'z'
}
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
func isNameStart(ch byte) bool {
	return isAlpha(ch) || ch == '_'
}
func isNameContinue(ch byte) bool {
	return isAlpha(ch) || isDigit(ch) || ch == '_'
}
