package parsec

import (
	"fmt"
	"regexp"
	"strings"
)

func Literal(str string) Parser {
	return func(s *Scanner) interface{} {
		s.Tracef("matching %q against %q", str, s.Buf[s.Loc.Pos:])
		if len(str) == 0 {
			return ""
		}

		c := len(str)
		end := s.Loc.Pos + c
		if end > len(s.Buf) {
			return &ParseFailure{
				loc: s.Loc,
				msg: fmt.Sprintf("Unexpected EOF while matching literal %q", str),
			}
		}

		if string(s.Buf[s.Loc.Pos:end]) != str {
			return &ParseFailure{
				loc: s.Loc,
				msg: fmt.Sprintf("Literal %q did not match", str),
			}
		}

		s.Advance(len(str))
		return str
	}
}

func SkipAny(chars string) Parser {
	return func(s *Scanner) interface{} {
		for !s.IsEOF() && strings.ContainsAny(s.Buf[s.Loc.Pos:s.Loc.Pos+1], chars) {
			s.Advance(1)
		}
		return nil
	}
}

func Match(re *regexp.Regexp) Parser {
	orig := re
	re, _ = regexp.Compile("^(?:" + re.String() + ")")
	return func(s *Scanner) interface{} {
		s.Tracef("matching %q against %q", orig, s.Buf[s.Loc.Pos:])
		loc := re.FindStringIndex(s.Buf[s.Loc.Pos:])
		if loc == nil {
			return &ParseFailure{
				loc: s.Loc,
				msg: fmt.Sprintf("Pattern %q did not match", orig.String()),
			}
		}
		loc[0] += s.Loc.Pos
		loc[1] += s.Loc.Pos
		s.Advance(loc[1] - loc[0])
		return s.Buf[loc[0]:loc[1]]
	}
}

func Submatches(re *regexp.Regexp) Parser {
	orig := re
	re, _ = regexp.Compile("^(?:" + re.String() + ")")
	return func(s *Scanner) interface{} {
		s.Tracef("matching %q against %q", orig, s.Buf[s.Loc.Pos:])
		subs := re.FindStringSubmatch(s.Buf[s.Loc.Pos:])
		if subs == nil {
			return &ParseFailure{
				loc: s.Loc,
				msg: fmt.Sprintf("Pattern %q did not match", orig.String()),
			}
		}
		s.Advance(len(subs[0]))
		return subs
	}
}

func Sequence(parsers ...Parser) Parser {
	return func(s *Scanner) interface{} {
		var result []interface{}
		for _, parser := range parsers {
			if !s.Parse(parser) {
				return s.LastFailure
			}
			result = append(result, s.Result)
		}
		return result
	}
}
