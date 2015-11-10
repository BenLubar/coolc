package ast

import (
	"bytes"
	"fmt"
	"go/token"
	"io"
	"os"
	"strconv"
	"unicode"
)

func (p *Program) Parse(f *token.File, opt Options, r *bytes.Reader) (haveErrors bool) {
	l := &lex{
		file: f,
		r:    r,

		program: p,

		opt: opt,
	}

	yyParse(l)

	return l.haveError
}

var (
	eofSentinel   = new(byte)
	errorSentinel = new(byte)
	idTokens      = map[string]int{
		"case":      CASE,
		"class":     CLASS,
		"def":       DEF,
		"else":      ELSE,
		"extends":   EXTENDS,
		"false":     FALSE,
		"if":        IF,
		"match":     MATCH,
		"new":       NEW,
		"null":      NULL,
		"override":  OVERRIDE,
		"super":     SUPER,
		"this":      THIS,
		"true":      TRUE,
		"var":       VAR,
		"while":     WHILE,
		"native":    NATIVE,
		"abstract":  ILLEGAL,
		"catch":     ILLEGAL,
		"do":        ILLEGAL,
		"final":     ILLEGAL,
		"finally":   ILLEGAL,
		"for":       ILLEGAL,
		"forSome":   ILLEGAL,
		"implicit":  ILLEGAL,
		"import":    ILLEGAL,
		"lazy":      ILLEGAL,
		"object":    ILLEGAL,
		"package":   ILLEGAL,
		"private":   ILLEGAL,
		"protected": ILLEGAL,
		"requires":  ILLEGAL,
		"return":    ILLEGAL,
		"sealed":    ILLEGAL,
		"throw":     ILLEGAL,
		"trait":     ILLEGAL,
		"try":       ILLEGAL,
		"type":      ILLEGAL,
		"val":       ILLEGAL,
		"with":      ILLEGAL,
		"yield":     ILLEGAL,
	}
)

type lex struct {
	file   *token.File
	r      *bytes.Reader
	offset int

	program   *Program
	haveError bool

	opt Options
}

func (l *lex) Lex(lvalue *yySymType) (tok int) {
	defer func() {
		if r := recover(); r != nil {
			if r == errorSentinel {
				lvalue.pos = l.Pos()
				tok = ILLEGAL
				return
			}
			if r == eofSentinel {
				tok = 0
				return
			}
			panic(r)
		}
	}()

	unexpected := false

	check := func(err error) {
		if err != nil {
			if err == io.EOF {
				if unexpected {
					l.Error("unexpected EOF")
					panic(errorSentinel)
				}
				panic(eofSentinel)
			}
			l.Error(err.Error())
			panic(errorSentinel)
		}
	}

	l.offset = -1
	offset, err := l.r.Seek(0, os.SEEK_CUR)
	check(err)
	l.offset = int(offset)

	r, err := l.r.ReadByte()
	check(err)

	for {
		offset, err = l.r.Seek(0, os.SEEK_CUR)
		check(err)
		// if r isn't ignored, our offset will be 1 byte too far. if it
		// is, we'll come back to this line.
		offset--
		l.offset = int(offset)

		if unicode.IsSpace(rune(r)) {
			r, err = l.r.ReadByte()
			check(err)
			continue
		}

		if r == '/' {
			r, err = l.r.ReadByte()
			switch r {
			case '/':
				for {
					r, err = l.r.ReadByte()
					check(err)
					if r == '\n' {
						r, err = l.r.ReadByte()
						check(err)
						break
					}
				}
				continue
			case '*':
				unexpected = true
				for {
					r, err = l.r.ReadByte()
					check(err)
					if r == '*' {
						r, err = l.r.ReadByte()
						check(err)
						if r == '/' {
							unexpected = false
							r, err = l.r.ReadByte()
							check(err)
							break
						}
						check(l.r.UnreadByte())
					}
				}
				continue
			default:
				check(l.r.UnreadByte())
				r = '/'
			}
		}

		break
	}

	unexpected = true

	validIdentifier := func(r byte) bool {
		return r == '_' || unicode.IsUpper(rune(r)) || unicode.IsLower(rune(r)) || (r >= '0' && r <= '9')
	}

	var buf []byte
	if unicode.IsUpper(rune(r)) {
		buf = append(buf, r)
		for {
			r, err = l.r.ReadByte()
			check(err)
			if validIdentifier(r) {
				buf = append(buf, r)
			} else {
				check(l.r.UnreadByte())
				lvalue.id = &Ident{
					Pos:  l.file.Pos(int(offset)),
					Name: string(buf),
				}
				return TYPEID
			}
		}
	}
	if unicode.IsLower(rune(r)) {
		buf = append(buf, r)
		for {
			r, err = l.r.ReadByte()
			check(err)
			if validIdentifier(r) {
				buf = append(buf, r)
			} else {
				check(l.r.UnreadByte())
				s := string(buf)
				if tok, ok := idTokens[s]; ok {
					lvalue.pos = l.file.Pos(int(offset))
					return tok
				}
				lvalue.id = &Ident{
					Pos:  l.file.Pos(int(offset)),
					Name: s,
				}
				return OBJECTID
			}
		}
	}
	if r == '0' {
		r, err = l.r.ReadByte()
		check(err)
		check(l.r.UnreadByte())

		if r >= '0' && r <= '9' {
			return ILLEGAL
		}

		lvalue.int = &IntLit{
			Pos: l.file.Pos(int(offset)),
			Int: 0,
		}
		return INTEGER
	}
	if r >= '1' && r <= '9' {
		buf = append(buf, r)

		for {
			r, err = l.r.ReadByte()
			check(err)
			if r >= '0' && r <= '9' {
				buf = append(buf, r)
			} else {
				check(l.r.UnreadByte())
				var n int64
				n, err = strconv.ParseInt(string(buf), 10, 32)
				check(err)
				lvalue.int = &IntLit{
					Pos: l.file.Pos(int(offset)),
					Int: int32(n),
				}
				return INTEGER
			}
		}
	}
	if r == '"' {
		r, err = l.r.ReadByte()
		check(err)
		if r == '"' {
			r, err = l.r.ReadByte()
			check(err)
			if r == '"' {
				for {
					r, err = l.r.ReadByte()
					check(err)
					buf = append(buf, r)

					if len(buf) >= 3 && buf[len(buf)-3] == '"' && buf[len(buf)-2] == '"' && buf[len(buf)-1] == '"' {
						lvalue.str = &StringLit{
							Pos: l.file.Pos(int(offset)),
							Str: string(buf[:len(buf)-3]),
						}
						return STRING
					}
				}
			}
			check(l.r.UnreadByte())
			lvalue.str = &StringLit{
				Pos: l.file.Pos(int(offset)),
				Str: "",
			}
			return STRING
		}

		for {
			switch r {
			case '\n':
				lvalue.pos = l.file.Pos(int(offset))
				return ILLEGAL

			case '"':
				lvalue.str = &StringLit{
					Pos: l.file.Pos(int(offset)),
					Str: string(buf),
				}
				return STRING

			case '\\':
				r, err = l.r.ReadByte()
				check(err)
				switch r {
				case '0':
					buf = append(buf, 0)
				case 'b':
					buf = append(buf, '\b')
				case 't':
					buf = append(buf, '\t')
				case 'n':
					buf = append(buf, '\n')
				case 'r':
					buf = append(buf, '\r')
				case 'f':
					buf = append(buf, '\f')
				case '"':
					buf = append(buf, '"')
				case '\\':
					buf = append(buf, '\\')
				default:
					lvalue.pos = l.file.Pos(int(offset))
					return ILLEGAL
				}

			default:
				buf = append(buf, r)
			}

			r, err = l.r.ReadByte()
			check(err)
		}
	}
	switch r {
	case '=':
		r, err = l.r.ReadByte()
		check(err)
		switch r {
		case '=':
			lvalue.pos = l.file.Pos(int(offset))
			return EQ
		case '>':
			lvalue.pos = l.file.Pos(int(offset))
			return ARROW
		default:
			check(l.r.UnreadByte())
			r = '='
		}
	case '<':
		r, err = l.r.ReadByte()
		check(err)
		switch r {
		case '=':
			lvalue.pos = l.file.Pos(int(offset))
			return LE
		default:
			check(l.r.UnreadByte())
			r = '<'
		}
	}
	lvalue.pos = l.file.Pos(l.offset)
	return int(r)
}

func (l *lex) Error(s string) {
	fmt.Fprintf(l.opt.Errors, "%v: %s\n", l.file.Position(l.Pos()), s)
	l.haveError = true
}

func (l *lex) Pos() token.Pos {
	return l.file.Pos(l.offset)
}
