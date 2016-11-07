package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/devnev/tiny-parsec"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	for {
		l, err := r.ReadString('\n')
		if l != "" {
			fmt.Printf("Processing %q\n", l)
			fmt.Printf("Got %+v\n", parseExpr(&parsec.Scanner{Buf: l /*, TraceTo: os.Stdout*/}))
		}
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func parseExpr(s *parsec.Scanner) interface{} {
	s.Trace("parsing expr")
	if s.Parse(parseSum) {
		return s.Result
	}
	return s.LastFailure
}

func parseSum(s *parsec.Scanner) interface{} {
	s.Trace("parsing sum")
	if !s.Parse(parseProduct) {
		return s.LastFailure
	}
	result := s.Result.(float64)
	for {
		var apply func(float64)
		if s.Parse(parsec.Literal("+")) {
			apply = func(v float64) { result = result + v }
		} else if s.Parse(parsec.Literal("-")) {
			apply = func(v float64) { result = result - v }
		} else {
			return result
		}
		if !s.Parse(parseProduct) {
			return s.LastFailure
		}
		apply(s.Result.(float64))
	}
}

func parseProduct(s *parsec.Scanner) interface{} {
	s.Trace("parsing product")
	if !s.Parse(parseValue) {
		return s.LastFailure
	}
	result := s.Result.(float64)
	for {
		var apply func(float64)
		if s.Parse(parsec.Literal("*")) {
			apply = func(v float64) { result = result * v }
		} else if s.Parse(parsec.Literal("/")) {
			apply = func(v float64) { result = result / v }
		} else {
			return result
		}
		if !s.Parse(parseValue) {
			return s.LastFailure
		}
		apply(s.Result.(float64))
	}
}

func parseValue(s *parsec.Scanner) interface{} {
	s.Parse(skipWs)
	s.Trace("parsing value")
	if s.Parse(parseNumber) {
		v, _ := strconv.ParseFloat(s.Result.(string), 64)
		s.Parse(skipWs)
		return v
	}
	if !s.Parse(parsec.Literal("(")) {
		return s.Fail("invalid value")
	}
	if !s.Parse(parseExpr) {
		return s.LastFailure
	}
	result := s.Result.(float64)
	if !s.Parse(parsec.Literal(")")) {
		return s.LastFailure
	}
	s.Parse(skipWs)
	return result
}

var parseNumber = parsec.Match(regexp.MustCompile("[0-9]+"))

var skipWs = parsec.SkipAny(" \t\n")
