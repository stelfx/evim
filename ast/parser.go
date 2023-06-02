package ast

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

type tokenType int

const (
	openParen tokenType = iota
	closeParen
	op
	constant
)

type token struct {
	typ   tokenType
	value string
}

type lexer struct {
	input  string
	start  int
	pos    int
	width  int
	tokens chan token
}

const eof rune = -1

type stateFunc func(*lexer) stateFunc

func BeginLexing(s string) Node {
	l := &lexer{input: s, tokens: make(chan token, 100)}
	go l.run()
	return parse(l.tokens, nil)
}

func (l *lexer) run() {
	for state := determineToken; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

func stringToNode(s string) Node {
	switch s {
	case "Clip":
		return NewOpClip()
	case "Negate":
		return NewOpNegate()
	case "*":
		return NewOpMult()
	case "+":
		return NewOpPlus()
	case "Ceil":
		return NewOpCeil()
	case "-":
		return NewOpMinus()
	case "/":
		return NewOpDiv()
	case "Square":
		return NewOpSquare()
	case "Lerp":
		return NewOpLerp()
	case "Sin":
		return NewOpSin()
	case "Cos":
		return NewOpCos()
	case "Floor":
		return NewOpFloor()
	case "Log":
		return NewOpLog()
	case "Wrap":
		return NewOpWrap()
	case "Abs":
		return NewOpAbs()
	case "Atan":
		return NewOpAtan()
	case "Noise":
		return NewOpNoise()
	case "FBM":
		return NewOpFBM()
	case "Turbulence":
		return NewOpTurbulence()
	case "Gamma":
		return NewOpGamma()
	case "X":
		return NewOpX()
	case "Y":
		return NewOpY()
	case "Picture":
		return NewOpPicture()
	case "Hypot":
		return NewOpHypot()
	default:
		panic("error in parser" + s)
	}
}

func parse(tokens chan token, parent Node) Node {

	for {
		token, ok := <-tokens
		if !ok {
			panic("no more tokens")

		}

		switch token.typ {
		case op:
			n := stringToNode(token.value)
			n.SetParent(parent)

			for i := range n.GetChildren() {
				n.GetChildren()[i] = parse(tokens, n)
			}
			return n

		case constant:
			n := NewOpConstant()
			n.SetParent(parent)
			v, err := strconv.ParseFloat(token.value, 32)
			if err != nil {
				panic(err)
			}
			n.value = float32(v)
			return n

		case closeParen, openParen:
			continue
		}
	}

	return nil
}

func determineToken(l *lexer) stateFunc {
	for {
		switch r := l.next(); {
		case isWhiteSpace(r):
			l.ignore()
		case r == '(':
			l.emit(openParen)
		case r == ')':
			l.emit(closeParen)
		case isStartNumber(r):
			return lexNumber
		case r == eof:
			return nil
		default:
			return lexOp
		}
	}
}

func lexOp(l *lexer) stateFunc {
	l.acceptRun("+=/*qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890")
	l.emit(op)
	return determineToken
}

func lexNumber(l *lexer) stateFunc {
	l.accept("-.")
	digits := "0123456789"
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}

	if l.input[l.start:l.pos] == "-" {
		l.emit(op)
	} else {
		l.emit(constant)
	}

	return determineToken
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {

	}
	l.backup()
}

func isWhiteSpace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\t' || r == '\r'
}

func isStartNumber(r rune) bool {
	return (r >= '0' && r <= '9') || r == '-' || r == '.'
}

func (l *lexer) emit(t tokenType) {
	l.tokens <- token{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) peek() (r rune) {
	r, _ = utf8.DecodeRuneInString(l.input[l.pos:])
	return r
}
