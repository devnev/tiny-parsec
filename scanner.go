package parsec

import (
	"fmt"
	"io"
)

type Scanner struct {
	TraceTo     io.Writer
	TraceLvl    int
	Buf         string
	Loc         Position
	Args        []interface{}
	Result      interface{}
	LastFailure *ParseFailure
}

func (s *Scanner) Trace(msg string) {
	if s.TraceTo != nil {
		for i := 0; i < s.TraceLvl; i++ {
			fmt.Fprint(s.TraceTo, "  ")
		}
		fmt.Fprintln(s.TraceTo, msg)
	}
}

func (s *Scanner) Tracef(format string, a ...interface{}) {
	if s.TraceTo != nil {
		s.Trace(fmt.Sprintf(format, a...))
	}
}

func (s *Scanner) IsEOF() bool {
	return s.Loc.Pos >= len(s.Buf)
}

func (s *Scanner) Copy() *Scanner {
	return &Scanner{
		TraceTo:  s.TraceTo,
		TraceLvl: s.TraceLvl + 1,
		Buf:      s.Buf,
		Loc:      s.Loc,
	}
}

func (s *Scanner) Advance(n int) {
	for ; n > 0; n -= 1 {
		s.Loc.Col += 1
		if s.Buf[s.Loc.Pos] == '\n' {
			s.Loc.Line += 1
			s.Loc.Col = 0
		}

		s.Loc.Pos += 1
	}
}

func (s *Scanner) subParse(parser Parser, args []interface{}) (*Scanner, interface{}, *ParseFailure) {
	sub := s.Copy()
	sub.Args = args
	result := parser(sub)
	if f, ok := result.(*ParseFailure); ok {
		s.Tracef("! %+v", f)
		return sub, result, f
	}
	s.Tracef("+ %T %#v", result, result)
	return sub, result, nil
}

func (s *Scanner) Parse(parser Parser, args ...interface{}) bool {
	sub, result, failure := s.subParse(parser, args)
	s.Result = result
	s.LastFailure = failure
	if failure == nil {
		s.Loc = sub.Loc
	}
	return failure == nil
}

func (s *Scanner) Peek(parser Parser, args ...interface{}) bool {
	_, _, failure := s.subParse(parser, args)
	return failure == nil
}

func (s *Scanner) Skip(parser Parser, args ...interface{}) bool {
	sub, _, failure := s.subParse(parser, args)
	if failure == nil {
		s.Loc = sub.Loc
	}
	return failure == nil
}

func (s *Scanner) Fail(msg string) *ParseFailure {
	return &ParseFailure{
		loc: s.Loc,
		msg: msg,
		sub: s.LastFailure,
	}
	return nil
}

func (s *Scanner) Failf(format string, a ...interface{}) *ParseFailure {
	return &ParseFailure{
		loc: s.Loc,
		msg: fmt.Sprintf(format, a...),
		sub: s.LastFailure,
	}
	return nil
}
