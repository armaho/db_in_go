package db_in_go

import (
	"errors"
	"strconv"
	"strings"
)

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

func (p *Parser) parseInt(out *Cell) error {
	start := p.pos
	end := p.pos + 1
	ch, ok := p.current(), true
	if ch == '-' || ch == '+' {
		ch, ok = p.advance()
		end += 1
	}
	if !isDigit(ch) {
		return errors.New("expected digit after +/-")
	}
	for ok && isDigit(ch) {
		ch, ok = p.advance()
		if ok {
			end += 1
		}
	}

	if !p.isEnd() && !isSeparator(p.current()) {
		return errors.New("expected seperator after num")
	}

	p.skipSpace()
	num, err := strconv.Atoi(p.buf[start:end])
	if err != nil {
		panic("cannot parse num")
	}
	out.Type = TypeI64
	out.I64 = int64(num)

	return nil
}

func (p *Parser) parseString(out *Cell) error {
	quote := p.buf[p.pos]
	cur := p.pos + 1
	for cur < len(p.buf) {
		ch := p.buf[cur]
		if ch == '\\' {
			cur++
			if cur < len(p.buf) && (p.buf[cur] == '"' || p.buf[cur] == '\'') {
				out.Str = append(out.Str, p.buf[cur])
				cur++
			} else {
				return errors.New("bad escape")
			}
		} else if ch == quote {
			out.Type = TypeStr
			p.pos = cur + 1
			return nil
		} else {
			out.Str = append(out.Str, p.buf[cur])
			cur++
		}
	}
	return errors.New("string is not terminated")
}

func (p *Parser) parseValue(out *Cell) error {
	if p.isEnd() {
		return errors.New("expect value")
	}
	ch := p.current()
	if ch == '"' || ch == '\'' {
		return p.parseString(out)
	} else if isDigit(ch) || ch == '-' || ch == '+' {
		return p.parseInt(out)
	} else {
		return errors.New("expect value")
	}
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
