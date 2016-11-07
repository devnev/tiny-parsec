package parsec

import (
	"fmt"
	"io"
)

// A Parser advances a scanner and returns either a *ParseFailure or a result object.
type Parser func(*Scanner) interface{}

type Position struct {
	Pos, Line, Col int
}

type ParseFailure struct {
	loc Position
	msg string
	sub *ParseFailure
}

func (f *ParseFailure) String() string {
	var sub string
	if f.sub != nil {
		sub = " because " + f.sub.String()
	}
	return fmt.Sprintf("at line %v, col %v: %v%v", f.loc.Line, f.loc.Col, f.msg, sub)
}

// ParseFirst runs the parser against the given input.
func ParseFirst(p Parser, s string) (interface{}, *ParseFailure) {
	r := p(&Scanner{Buf: s})
	if f, ok := r.(*ParseFailure); ok {
		return nil, f
	}
	return r, nil
}

// Parse runs the parser against the given input and checks that all input was consumed.
func Parse(p Parser, s string) (interface{}, *ParseFailure) {
	sc := &Scanner{Buf: s}
	r := p(sc)
	if f, ok := r.(*ParseFailure); ok {
		return nil, f
	}
	if !sc.IsEOF() {
		return nil, sc.Fail("expected end of stream")
	}
	return r, nil
}

// TraceParse runs the parser against the given input while printing tracing information to a writer.
func TraceParse(p Parser, s string, out io.Writer) (interface{}, *ParseFailure) {
	r := p(&Scanner{Buf: s, TraceTo: out})
	if f, ok := r.(*ParseFailure); ok {
		return nil, f
	}
	return r, nil
}
