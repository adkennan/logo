package main

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

const comment rune = ';'
const escape rune = '\\'
const listStart rune = '['
const listEnd rune = ']'
const literalStart rune = '"'
const newLine rune = '\n'
const thingStart rune = ':'

var wordSeparators = []rune{' ', '\t', newLine, comment, listEnd}
var infixOpChars = []rune{'+', '-', '*', '/', '<', '>', '=', '(', ')'}
var listSeparators = []rune{' ', '\t', newLine, comment}

func ParseString(text string) (n Node, err error) {
	r := strings.NewReader(text)
	return Parse(r)
}

func Parse(r io.Reader) (n Node, err error) {

	l := 1
	c := 1
	rr := bufio.NewReader(r)
	var pn Node = nil
	var nn Node = nil
	for err == nil {
		nn, err = parse(rr, &l, &c)

		if nn != nil {
			if pn != nil {
				pn.addNode(nn)
				for nn.next() != nil {
					nn = nn.next()
				}
			} else {
				n = nn
			}
			pn = nn
		}
	}

	if err == io.EOF {
		err = nil
	}

	return n, err

}

func parse(r *bufio.Reader, line, col *int) (n Node, err error) {

	err = readSeparator(r, List, line, col)
	if err != nil {
		return
	}

	c, _, err := r.ReadRune()
	checkNewline(c, line, col)
	switch c {
	case comment:
		err = readComment(r, line, col)
	case listStart:
		n, err = readList(r, line, col)
	case listEnd:
		r.UnreadRune()
	default:
		n, err = readWord(r, line, col)
	}

	return
}

func isSeparator(c rune, nt NodeType) bool {

	if nt == Word {
		for _, s := range wordSeparators {
			if c == s {
				return true
			}
		}
	} else {
		for _, s := range listSeparators {
			if c == s {
				return true
			}
		}
	}
	return false
}

func isOperator(c rune) bool {
	for _, s := range infixOpChars {
		if s == c {
			return true
		}
	}
	return false
}

func checkNewline(c rune, line, col *int) {
	if c == newLine {
		(*line)++
		(*col) = 1
	} else {
		(*col)++
	}
}

func readSeparator(r *bufio.Reader, nt NodeType, line, col *int) (err error) {

	c, _, err := r.ReadRune()
	checkNewline(c, line, col)
	for err == nil && isSeparator(c, nt) {
		c, _, err = r.ReadRune()
		checkNewline(c, line, col)
	}

	if err == nil || err == io.EOF {
		r.UnreadRune()
	}
	return err
}

func readComment(r *bufio.Reader, line, col *int) (err error) {

	c, _, err := r.ReadRune()
	checkNewline(c, line, col)
	for err == nil && c != newLine {
		c, _, err = r.ReadRune()
		checkNewline(c, line, col)
	}
	return err
}

func readList(r *bufio.Reader, line, col *int) (n Node, err error) {

	var fn Node = nil
	var pn Node = nil
	var nn Node = nil
	var closed bool = false
	for {
		nn, err = parse(r, line, col)
		if err != nil {
			break
		}
		if nn != nil {
			if pn != nil {
				pn.addNode(nn)
			}
			pn = nn
			if fn == nil {
				fn = nn
			}

		}
		err = readSeparator(r, List, line, col)
		if err != nil {
			break
		}

		c, _, err := r.ReadRune()
		if err != nil {
			break
		}

		checkNewline(c, line, col)

		if c == listEnd {
			closed = true
			break
		}
		r.UnreadRune()
	}

	if err == io.EOF && !closed {
		err = io.ErrUnexpectedEOF
	}

	if err == nil {
		return newListNode(*line, *col, fn), nil
	}
	return nil, err
}

func isNegativeNumber(pc, c rune, r *bufio.Reader) bool {

	if c != '-' {
		return false
	}

	if pc != 0 && !isSeparator(pc, Word) {
		return false
	}

	nc, _, _ := r.ReadRune()
	r.UnreadRune()
	return unicode.IsDigit(nc)
}

func readWord(r *bufio.Reader, line, col *int) (n Node, err error) {

	r.UnreadRune()

	chars := make([]rune, 0, 4)
	escaped := false
	isLiteral := false
	var pc rune

	for {
		c, _, err := r.ReadRune()
		if err != nil {
			break
		}
		if !escaped && isOperator(c) && !isNegativeNumber(pc, c, r) {
			if len(chars) == 0 {
				chars = append(chars, c)
				if c == '<' || c == '>' {
					nc, _, err := r.ReadRune()
					if err != nil {
						break
					}
					if (c == '<' && (nc == '>' || nc == '=')) ||
						(c == '>' && nc == '=') {
						chars = append(chars, nc)

					} else {
						r.UnreadRune()
					}
					break
				}
			} else {
				r.UnreadRune()
				break
			}
		} else {

			if len(chars) == 0 && (c == literalStart || unicode.IsDigit(c) || c == '-') {

				isLiteral = true
			}

			checkNewline(c, line, col)

			if !escaped {
				if isSeparator(c, Word) {
					r.UnreadRune()
					break
				}

				if c == escape {
					escaped = true
					continue
				}
			} else {
				escaped = false
			}

			if len(chars) > 0 || c != literalStart {
				chars = append(chars, c)
				pc = c
			}
		}
	}

	if len(chars) > 0 {
		return newWordNode(*line, *col, string(chars), isLiteral), err
	}
	return nil, err
}
