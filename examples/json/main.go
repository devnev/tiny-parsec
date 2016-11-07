package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/devnev/tiny-parsec"
)

func main() {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(parsec.Parse(parseObject, string(b)))
}

func parseObject(s *parsec.Scanner) interface{} {
	s.Parse(skipWs)
	if !s.Parse(parsec.Literal("{")) {
		return s.LastFailure
	}
	s.Parse(skipWs)
	obj := map[string]interface{}{}
	if s.Parse(parsec.Literal("}")) {
		return obj
	}
	for {
		s.Parse(skipWs)
		if !s.Parse(parseString) {
			return s.LastFailure
		}
		key := s.Result.(string)
		if _, ok := obj[key]; ok {
			return s.Failf("duplicate key %q", key)
		}
		s.Parse(skipWs)
		if !s.Parse(parsec.Literal(":")) {
			return s.LastFailure
		}
		if !s.Parse(parseValue) {
			return s.LastFailure
		}
		obj[key] = s.Result
		s.Parse(skipWs)
		if !s.Parse(parsec.Literal(",")) {
			break
		}
	}
	if !s.Parse(parsec.Literal("}")) {
		return s.LastFailure
	}
	s.Parse(skipWs)
	return obj
}

func parseArray(s *parsec.Scanner) interface{} {
	s.Parse(skipWs)
	if !s.Parse(parsec.Literal("[")) {
		return s.LastFailure
	}
	arr := []interface{}{}
	s.Parse(skipWs)
	if s.Parse(parsec.Literal("]")) {
		return arr
	}
	for {
		if !s.Parse(parseValue) {
			return s.LastFailure
		}
		arr = append(arr, s.Result)
		s.Parse(skipWs)
		if !s.Parse(parsec.Literal(",")) {
			break
		}
	}
	if !s.Parse(parsec.Literal("]")) {
		return s.LastFailure
	}
	return arr
}

func parseValue(s *parsec.Scanner) interface{} {
	s.Parse(skipWs)
	if s.Peek(parsec.Literal("\"")) {
		s.Parse(parseString)
		return s.Result
	}
	if s.Peek(parsec.Literal("{")) {
		s.Parse(parseObject)
		return s.Result
	}
	if s.Peek(parsec.Literal("[")) {
		s.Parse(parseArray)
		return s.Result
	}
	if s.Parse(parseNumber) {
		return s.Result
	}
	if s.Parse(parsec.Literal("true")) {
		return true
	}
	if s.Parse(parsec.Literal("false")) {
		return false
	}
	if s.Parse(parsec.Literal("null")) {
		return nil
	}
	return s.Fail("unable to parse value")
}

func parseString(s *parsec.Scanner) interface{} {
	if !s.Parse(parsec.Literal("\"")) {
		return s.LastFailure
	}
	if !s.Parse(matchStringChars) {
		return s.Fail("bug")
	}
	str, err := processStrChars(s.Result.(string))
	if err != nil {
		return s.Fail(err.Error())
	}
	if !s.Parse(parsec.Literal("\"")) {
		return s.Fail("invalid character in string literal")
	}
	return str
}

func parseNumber(s *parsec.Scanner) interface{} {
	if !s.Parse(matchNumber) {
		return s.Fail("not a valid number")
	}
	v, err := strconv.ParseFloat(s.Result.(string), 64)
	if err != nil {
		return s.Fail(err.Error())
	}
	return v
}

func processStrChars(str string) (string, error) {
	src := strings.NewReader(str)
	buf := bytes.Buffer{}
	for {
		r, _, err := src.ReadRune()
		if err == io.EOF {
			return buf.String(), nil
		}
		if r != '\\' {
			buf.WriteRune(r)
			continue
		}
		r, _, _ = src.ReadRune()
		switch r {
		case 'b':
			r = '\b'
		case 'f':
			r = '\f'
		case 'n':
			r = '\n'
		case 'r':
			r = '\r'
		case 't':
			r = '\t'
		case 'u':
			rs := [4]byte{}
			for i := 0; i < len(rs); i++ {
				rs[i], _ = src.ReadByte()
			}
			s := string(rs[:])
			v, err := strconv.ParseInt(s, 16, 64)
			if err != nil {
				return "", err
			}
			if v < 0 || v > int64(unicode.MaxRune) {
				return "", fmt.Errorf("invalid unicode character %q: codepoint %v out of range", s, v)
			}
			r = rune(v)
		default:
			return "", errors.New("bug")
		}
		buf.WriteRune(r)
	}
}

var matchNumber = parsec.Match(regexp.MustCompile(`(-?(?:0|[1-9][0-9]*))(?:\.([0-9]+))?([eE][+-]?[0-9]+)?`))

var matchStringChars = parsec.Match(regexp.MustCompile(`(?:\\["\\/bfnrt]|\\u[0-9a-fA-F]{4}|[^\\"\p{Cc}])*`))

var skipWs = parsec.Match(regexp.MustCompile(`[\p{Zs}\t\r\n]*`))
