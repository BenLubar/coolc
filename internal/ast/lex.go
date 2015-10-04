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

func (p *Program) Parse(f *token.File, r *bytes.Reader) (haveErrors bool) {
	l := &lex{
		file: f,
		r:    r,

		program: p,
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

	r, _, err := l.r.ReadRune()
	check(err)

	for {
		offset, err = l.r.Seek(0, os.SEEK_CUR)
		check(err)
		l.offset = int(offset)

		if unicode.IsSpace(r) {
			r, _, err = l.r.ReadRune()
			check(err)
			continue
		}

		if r == '/' {
			r, _, err = l.r.ReadRune()
			switch r {
			case '/':
				for {
					r, _, err = l.r.ReadRune()
					check(err)
					if r == '\n' {
						r, _, err = l.r.ReadRune()
						check(err)
						break
					}
				}
				continue
			case '*':
				unexpected = true
				for {
					r, _, err = l.r.ReadRune()
					check(err)
					if r == '*' {
						r, _, err = l.r.ReadRune()
						check(err)
						if r == '/' {
							unexpected = false
							r, _, err = l.r.ReadRune()
							check(err)
							break
						}
						check(l.r.UnreadRune())
					}
				}
				continue
			default:
				check(l.r.UnreadRune())
				r = '/'
			}
		}

		break
	}

	unexpected = true

	validIdentifier := func(r rune) bool {
		return r == '_' || unicode.IsUpper(r) || unicode.IsLower(r) || (r >= '0' && r <= '9')
	}

	var buf []rune
	if unicode.IsUpper(r) {
		buf = append(buf, r)
		for {
			r, _, err = l.r.ReadRune()
			check(err)
			if validIdentifier(r) {
				buf = append(buf, r)
			} else {
				check(l.r.UnreadRune())
				lvalue.id = &Ident{
					Pos:  l.file.Pos(int(offset)),
					Name: string(buf),
				}
				return TYPEID
			}
		}
	}
	if unicode.IsLower(r) {
		buf = append(buf, r)
		for {
			r, _, err = l.r.ReadRune()
			check(err)
			if validIdentifier(r) {
				buf = append(buf, r)
			} else {
				check(l.r.UnreadRune())
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
		r, _, err = l.r.ReadRune()
		check(err)
		check(l.r.UnreadRune())

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
			r, _, err = l.r.ReadRune()
			check(err)
			if r >= '0' && r <= '9' {
				buf = append(buf, r)
			} else {
				check(l.r.UnreadRune())
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
		r, _, err = l.r.ReadRune()
		check(err)
		if r == '"' {
			r, _, err = l.r.ReadRune()
			check(err)
			if r == '"' {
				for {
					r, _, err = l.r.ReadRune()
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
			check(l.r.UnreadRune())
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
				r, _, err = l.r.ReadRune()
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

			r, _, err = l.r.ReadRune()
			check(err)
		}
	}
	switch r {
	case '=':
		r, _, err = l.r.ReadRune()
		check(err)
		switch r {
		case '=':
			lvalue.pos = l.file.Pos(int(offset))
			return EQ
		case '>':
			lvalue.pos = l.file.Pos(int(offset))
			return ARROW
		default:
			check(l.r.UnreadRune())
			r = '='
		}
	case '<':
		r, _, err = l.r.ReadRune()
		check(err)
		switch r {
		case '=':
			lvalue.pos = l.file.Pos(int(offset))
			return LE
		default:
			check(l.r.UnreadRune())
			r = '<'
		}
	}
	lvalue.pos = l.file.Pos(l.offset)
	return int(r)
}

func (l *lex) Error(s string) {
	l.haveError = true
	fmt.Printf("%v: %s\n", l.file.Position(l.Pos()), s)
}

func (l *lex) Pos() token.Pos {
	return l.file.Pos(l.offset)
}
