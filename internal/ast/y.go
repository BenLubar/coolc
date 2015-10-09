//go:generate go tool yacc -l syntax.y

package ast

import __yyfmt__ "fmt"

import "go/token"

func init() {
	yyErrorVerbose = true
}

type yySymType struct {
	yys int
	pos token.Pos

	cl  *Class
	cls []*Class
	ft  Feature
	fts []Feature
	fm  *Formal
	fms []*Formal
	ext *Extends
	exp Expr
	act []Expr
	ca  *Case
	cas []*Case

	id  *Ident
	int *IntLit
	bin *BoolLit
	str *StringLit
}

const CLASS = 57346
const EXTENDS = 57347
const NATIVE = 57348
const VAR = 57349
const DEF = 57350
const OVERRIDE = 57351
const SUPER = 57352
const NEW = 57353
const ELSE = 57354
const NULL = 57355
const THIS = 57356
const ARROW = 57357
const CASE = 57358
const TRUE = 57359
const FALSE = 57360
const ILLEGAL = 57361
const TYPEID = 57362
const OBJECTID = 57363
const INTEGER = 57364
const STRING = 57365
const IF = 57366
const WHILE = 57367
const MATCH = 57368
const LE = 57369
const EQ = 57370

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"CLASS",
	"EXTENDS",
	"NATIVE",
	"VAR",
	"DEF",
	"OVERRIDE",
	"SUPER",
	"NEW",
	"ELSE",
	"NULL",
	"THIS",
	"ARROW",
	"CASE",
	"TRUE",
	"FALSE",
	"ILLEGAL",
	"'('",
	"TYPEID",
	"OBJECTID",
	"INTEGER",
	"STRING",
	"'='",
	"IF",
	"WHILE",
	"MATCH",
	"LE",
	"'<'",
	"EQ",
	"'+'",
	"'-'",
	"'*'",
	"'/'",
	"'!'",
	"'.'",
	"')'",
	"'{'",
	"'}'",
	"';'",
	"':'",
	"','",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 70
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 276

var yyAct = [...]int{

	56, 13, 55, 54, 115, 33, 42, 43, 126, 46,
	50, 87, 62, 51, 52, 12, 45, 140, 37, 47,
	48, 138, 40, 41, 108, 35, 18, 105, 86, 39,
	53, 83, 38, 116, 82, 44, 97, 19, 144, 74,
	75, 142, 135, 31, 32, 30, 81, 125, 80, 70,
	63, 64, 65, 68, 69, 66, 67, 127, 71, 151,
	118, 61, 150, 89, 90, 91, 92, 93, 94, 95,
	96, 11, 71, 99, 35, 27, 25, 101, 102, 100,
	66, 67, 57, 71, 78, 42, 43, 107, 46, 50,
	113, 73, 51, 52, 136, 45, 72, 37, 47, 48,
	147, 40, 41, 124, 130, 35, 14, 103, 39, 29,
	122, 38, 98, 129, 44, 85, 60, 59, 35, 128,
	132, 133, 35, 131, 21, 137, 146, 134, 139, 70,
	63, 64, 65, 68, 69, 66, 67, 145, 71, 20,
	58, 123, 84, 155, 149, 148, 109, 42, 43, 79,
	46, 50, 154, 153, 51, 52, 156, 45, 121, 37,
	47, 48, 22, 40, 41, 5, 117, 104, 88, 77,
	39, 42, 43, 38, 46, 50, 44, 76, 51, 52,
	24, 45, 6, 37, 47, 48, 116, 40, 41, 65,
	68, 69, 66, 67, 39, 71, 152, 38, 141, 32,
	44, 70, 63, 64, 65, 68, 69, 66, 67, 10,
	71, 120, 70, 63, 64, 65, 68, 69, 66, 67,
	143, 71, 119, 70, 63, 64, 65, 68, 69, 66,
	67, 9, 71, 106, 110, 16, 70, 63, 64, 65,
	68, 69, 66, 67, 17, 71, 70, 63, 64, 65,
	68, 69, 66, 67, 4, 71, 68, 69, 66, 67,
	1, 71, 49, 114, 34, 36, 15, 112, 111, 8,
	7, 28, 26, 23, 3, 2,
}
var yyPact = [...]int{

	-1000, -1000, 250, -1000, 144, 162, 202, 33, -28, -1000,
	84, 230, 202, -1000, -16, -2, 118, -1000, 141, -1000,
	160, -1000, -1000, 36, 161, -1000, -11, 75, -1000, -1000,
	191, 95, 94, 23, -31, 218, -1000, 71, 161, 161,
	157, 149, 47, 128, 75, -4, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -9, -1000, 101, 93, -1000, -14,
	148, -1000, 161, 161, 161, 161, 161, 161, 161, 161,
	-3, 90, 161, 161, 35, 35, 161, 161, 85, 147,
	-13, 195, -1000, -1000, 75, -18, 125, 228, 84, 218,
	158, 158, 224, 35, 35, 46, 46, 170, 146, 218,
	22, 184, 173, 138, 161, -1000, -1000, -1000, 120, 78,
	-1000, 9, -35, -1000, 17, -1000, 91, 161, -1000, 161,
	161, 161, 4, 69, 161, -21, 84, -1000, -1000, -25,
	183, 3, 208, 218, 0, -1000, 161, 218, 105, -1000,
	79, 75, -1000, 161, -1000, 21, 34, 181, -1000, 218,
	75, 137, 75, -1000, 218, -1000, -1000,
}
var yyPgo = [...]int{

	0, 275, 274, 273, 272, 271, 109, 270, 269, 268,
	267, 231, 1, 266, 3, 2, 0, 265, 5, 264,
	263, 4, 262, 260,
}
var yyR1 = [...]int{

	0, 23, 1, 1, 2, 13, 13, 13, 3, 3,
	4, 4, 4, 4, 5, 5, 6, 6, 18, 18,
	19, 19, 14, 14, 15, 15, 15, 7, 7, 8,
	8, 11, 9, 9, 10, 10, 12, 16, 16, 16,
	16, 16, 16, 16, 16, 16, 16, 16, 16, 16,
	16, 16, 17, 17, 17, 17, 17, 17, 17, 17,
	17, 17, 17, 17, 22, 22, 20, 20, 21, 21,
}
var yyR2 = [...]int{

	0, 1, 0, 2, 9, 0, 5, 2, 0, 3,
	3, 1, 1, 2, 6, 4, 9, 9, 0, 1,
	1, 3, 0, 1, 1, 8, 3, 0, 1, 1,
	3, 2, 0, 1, 1, 3, 3, 1, 3, 2,
	2, 7, 5, 3, 3, 3, 3, 3, 3, 3,
	5, 6, 4, 6, 5, 3, 3, 1, 2, 1,
	1, 1, 1, 1, 1, 1, 1, 2, 6, 4,
}
var yyChk = [...]int{

	-1000, -23, -1, -2, 4, 21, 20, -7, -8, -11,
	7, 38, 43, -12, 22, -13, 5, -11, 42, 39,
	21, 6, 21, -3, 20, 40, -4, 39, -5, -6,
	9, 7, 8, -18, -19, -16, -17, 22, 36, 33,
	26, 27, 10, 11, 39, 20, 13, 23, 24, -22,
	14, 17, 18, 41, -14, -15, -16, 7, -6, 22,
	22, 38, 43, 29, 30, 31, 34, 35, 32, 33,
	28, 37, 25, 20, -16, -16, 20, 20, 37, 21,
	-14, -16, 38, 40, 41, 22, 42, 25, 20, -16,
	-16, -16, -16, -16, -16, -16, -16, 39, 22, -16,
	-18, -16, -16, 22, 20, 40, 38, -15, 42, 21,
	6, -9, -10, -12, -20, -21, 16, 20, 38, 38,
	38, 20, -18, 21, 25, 38, 43, 40, -21, 22,
	13, -18, -16, -16, -18, 38, 25, -16, 42, -12,
	42, 15, 38, 12, 38, -16, 21, 21, -14, -16,
	41, 25, 15, -15, -16, 6, -14,
}
var yyDef = [...]int{

	2, -2, 1, 3, 0, 0, 27, 0, 28, 29,
	0, 5, 0, 31, 0, 0, 0, 30, 0, 8,
	0, 7, 36, 0, 18, 4, 0, 22, 11, 12,
	0, 0, 0, 0, 19, 20, 37, 59, 0, 0,
	0, 0, 0, 0, 22, 0, 57, 60, 61, 62,
	63, 64, 65, 9, 0, 23, 24, 0, 13, 0,
	0, 6, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 18, 39, 40, 0, 0, 0, 0,
	0, 0, 58, 10, 0, 0, 0, 0, 32, 21,
	43, 44, 45, 46, 47, 48, 49, 0, 0, 38,
	0, 0, 0, 0, 18, 55, 56, 26, 0, 0,
	15, 0, 33, 34, 0, 66, 0, 18, 52, 0,
	0, 18, 0, 0, 0, 0, 0, 50, 67, 0,
	0, 0, 0, 42, 0, 54, 0, 14, 0, 35,
	0, 22, 51, 0, 53, 0, 0, 0, 69, 41,
	0, 0, 22, 25, 16, 17, 68,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 36, 3, 3, 3, 3, 3, 3,
	20, 38, 34, 32, 43, 33, 37, 35, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 42, 41,
	30, 25, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 39, 3, 40,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 21, 22,
	23, 24, 26, 27, 28, 29, 31,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lookahead func() int
}

func (p *yyParserImpl) Lookahead() int {
	return p.lookahead()
}

func yyNewParser() yyParser {
	p := &yyParserImpl{
		lookahead: func() int { return -1 },
	}
	return p
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := make([]yySymType, yyMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yychar := -1
	yytoken := -1 // yychar translated into internal numbering
	yyrcvr.lookahead = func() int { return yychar }
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yychar = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yychar < 0 {
		yychar, yytoken = yylex1(yylex, &yylval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yychar = -1
		yytoken = -1
		yyVAL = yylval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yychar < 0 {
			yychar, yytoken = yylex1(yylex, &yylval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yychar = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			p := yylex.(*lex).program
			p.Classes = append(p.Classes, yyDollar[1].cls...)
		}
	case 2:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.cls = nil
		}
	case 3:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.cls = append(yyDollar[1].cls, yyDollar[2].cl)
		}
	case 4:
		yyDollar = yyS[yypt-9 : yypt+1]
		{
			yyVAL.cl = &Class{
				Type:     yyDollar[2].id,
				Formals:  yyDollar[4].fms,
				Extends:  yyDollar[6].ext,
				Features: yyDollar[8].fts,
			}
		}
	case 5:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.ext = &Extends{
				Type: &Ident{
					Name: "Any",
					Pos:  token.NoPos,
				},
				Args: nil,
			}
		}
	case 6:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.ext = &Extends{
				Type: yyDollar[2].id,
				Args: yyDollar[4].act,
			}
		}
	case 7:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.ext = &Extends{
				Type: &Ident{
					Name: "native",
					Pos:  yyDollar[2].pos,
				},
				Args: nil,
			}
		}
	case 8:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.fts = nil
		}
	case 9:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.fts = append(yyDollar[1].fts, yyDollar[2].ft)
		}
	case 10:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.ft = &Init{
				Expr: yyDollar[2].exp,
			}
		}
	case 11:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.ft = yyDollar[1].ft
		}
	case 12:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.ft = yyDollar[1].ft
		}
	case 13:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.ft = yyDollar[2].ft
			yyVAL.ft.(*Method).Override = true
		}
	case 14:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.ft = &Attribute{
				Name: yyDollar[2].id,
				Type: yyDollar[4].id,
				Init: yyDollar[6].exp,
			}
		}
	case 15:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.ft = &Attribute{
				Name: yyDollar[2].id,
				Type: &Ident{
					Name: "native",
					Pos:  yyDollar[4].pos,
				},
				Init: &NativeExpr{
					Pos: yyDollar[4].pos,
				},
			}
		}
	case 16:
		yyDollar = yyS[yypt-9 : yypt+1]
		{
			yyVAL.ft = &Method{
				Name: yyDollar[2].id,
				Args: yyDollar[4].fms,
				Type: yyDollar[7].id,
				Body: yyDollar[9].exp,
			}
		}
	case 17:
		yyDollar = yyS[yypt-9 : yypt+1]
		{
			yyVAL.ft = &Method{
				Name: yyDollar[2].id,
				Args: yyDollar[4].fms,
				Type: yyDollar[7].id,
				Body: &NativeExpr{
					Pos: yyDollar[9].pos,
				},
			}
		}
	case 18:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.act = nil
		}
	case 19:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.act = yyDollar[1].act
		}
	case 20:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.act = []Expr{yyDollar[1].exp}
		}
	case 21:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.act = append(yyDollar[1].act, yyDollar[3].exp)
		}
	case 22:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.exp = &UnitExpr{
				Pos: yylex.(*lex).Pos(),
			}
		}
	case 23:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.exp = yyDollar[1].exp
		}
	case 24:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.exp = yyDollar[1].exp
		}
	case 25:
		yyDollar = yyS[yypt-8 : yypt+1]
		{
			yyVAL.exp = &VarExpr{
				Name: yyDollar[2].id,
				Type: yyDollar[4].id,
				Init: yyDollar[6].exp,
				Body: yyDollar[8].exp,
			}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.exp = &ChainExpr{
				Pre:  yyDollar[1].exp,
				Expr: yyDollar[3].exp,
			}
		}
	case 27:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.fms = nil
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.fms = yyDollar[1].fms
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.fms = []*Formal{yyDollar[1].fm}
		}
	case 30:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.fms = append(yyDollar[1].fms, yyDollar[3].fm)
		}
	case 31:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.fm = yyDollar[2].fm
		}
	case 32:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.fms = nil
		}
	case 33:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.fms = yyDollar[1].fms
		}
	case 34:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.fms = []*Formal{yyDollar[1].fm}
		}
	case 35:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.fms = append(yyDollar[1].fms, yyDollar[3].fm)
		}
	case 36:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.fm = &Formal{
				Name: yyDollar[1].id,
				Type: yyDollar[3].id,
			}
		}
	case 37:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.exp = yyDollar[1].exp
		}
	case 38:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.exp = &AssignExpr{
				Name: yyDollar[1].id,
				Expr: yyDollar[3].exp,

				Unit: &Ident{
					Name: "Unit",
					Pos:  yyDollar[2].pos,
				},
			}
		}
	case 39:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.exp = &NotExpr{
				Expr: yyDollar[2].exp,

				Boolean: &Ident{
					Name: "Boolean",
					Pos:  yyDollar[1].pos,
				},
			}
		}
	case 40:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.exp = &NegativeExpr{
				Expr: yyDollar[2].exp,

				Int: &Ident{
					Name: "Int",
					Pos:  yyDollar[1].pos,
				},
			}
		}
	case 41:
		yyDollar = yyS[yypt-7 : yypt+1]
		{
			yyVAL.exp = &IfExpr{
				Cond: yyDollar[3].exp,
				Then: yyDollar[5].exp,
				Else: yyDollar[7].exp,

				Boolean: &Ident{
					Name: "Boolean",
					Pos:  yyDollar[1].pos,
				},
			}
		}
	case 42:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.exp = &WhileExpr{
				Cond: yyDollar[3].exp,
				Body: yyDollar[5].exp,

				Boolean: &Ident{
					Name: "Boolean",
					Pos:  yyDollar[1].pos,
				},
				Unit: &Ident{
					Name: "Unit",
					Pos:  yyDollar[1].pos,
				},
			}
		}
	case 43:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.exp = &LessOrEqualExpr{
				Pos:   yyDollar[2].pos,
				Left:  yyDollar[1].exp,
				Right: yyDollar[3].exp,

				Boolean: &Ident{
					Name: "Boolean",
					Pos:  yyDollar[2].pos,
				},
				Int: &Ident{
					Name: "Int",
					Pos:  yyDollar[2].pos,
				},
			}
		}
	case 44:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.exp = &LessThanExpr{
				Pos:   yyDollar[2].pos,
				Left:  yyDollar[1].exp,
				Right: yyDollar[3].exp,

				Boolean: &Ident{
					Name: "Boolean",
					Pos:  yyDollar[2].pos,
				},
				Int: &Ident{
					Name: "Int",
					Pos:  yyDollar[2].pos,
				},
			}
		}
	case 45:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.exp = &DynamicCallExpr{
				Recv: yyDollar[1].exp,
				Name: &Ident{
					Name: "equals",
					Pos:  yyDollar[2].pos,
				},
				Args: []Expr{
					yyDollar[3].exp,
				},
			}
		}
	case 46:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.exp = &MultiplyExpr{
				Pos:   yyDollar[2].pos,
				Left:  yyDollar[1].exp,
				Right: yyDollar[3].exp,

				Int: &Ident{
					Name: "Int",
					Pos:  yyDollar[2].pos,
				},
			}
		}
	case 47:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.exp = &DivideExpr{
				Pos:   yyDollar[2].pos,
				Left:  yyDollar[1].exp,
				Right: yyDollar[3].exp,

				Int: &Ident{
					Name: "Int",
					Pos:  yyDollar[2].pos,
				},
			}
		}
	case 48:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.exp = &AddExpr{
				Pos:   yyDollar[2].pos,
				Left:  yyDollar[1].exp,
				Right: yyDollar[3].exp,

				Int: &Ident{
					Name: "Int",
					Pos:  yyDollar[2].pos,
				},
			}
		}
	case 49:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.exp = &SubtractExpr{
				Pos:   yyDollar[2].pos,
				Left:  yyDollar[1].exp,
				Right: yyDollar[3].exp,

				Int: &Ident{
					Name: "Int",
					Pos:  yyDollar[2].pos,
				},
			}
		}
	case 50:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.exp = &MatchExpr{
				Pos:   yyDollar[2].pos,
				Left:  yyDollar[1].exp,
				Cases: yyDollar[4].cas,
			}
		}
	case 51:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.exp = &DynamicCallExpr{
				Recv: yyDollar[1].exp,
				Name: yyDollar[3].id,
				Args: yyDollar[5].act,
			}
		}
	case 52:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.exp = &DynamicCallExpr{
				Recv: &ThisExpr{
					Pos: yyDollar[1].id.Pos,
				},
				Name: yyDollar[1].id,
				Args: yyDollar[3].act,
			}
		}
	case 53:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.exp = &SuperCallExpr{
				Pos:  yyDollar[1].pos,
				Name: yyDollar[3].id,
				Args: yyDollar[5].act,
			}
		}
	case 54:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.exp = &StaticCallExpr{
				Recv: &AllocExpr{
					Type: yyDollar[2].id,
				},
				// $2 refers to a type, so we need to make a copy
				// because this one is the constructor method:
				Name: &Ident{
					Name: yyDollar[2].id.Name,
					Pos:  yyDollar[2].id.Pos,
				},
				Args: yyDollar[4].act,
			}
		}
	case 55:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.exp = yyDollar[2].exp
		}
	case 56:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.exp = yyDollar[2].exp
		}
	case 57:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.exp = &NullExpr{
				Pos: yyDollar[1].pos,
			}
		}
	case 58:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.exp = &UnitExpr{
				Pos: yyDollar[1].pos,
			}
		}
	case 59:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.exp = &NameExpr{
				Name: yyDollar[1].id,
			}
		}
	case 60:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.exp = &IntExpr{
				Lit: yyDollar[1].int,
			}
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.exp = &StringExpr{
				Lit: yyDollar[1].str,
			}
		}
	case 62:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.exp = &BoolExpr{
				Lit: yyDollar[1].bin,
			}
		}
	case 63:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.exp = &ThisExpr{
				Pos: yyDollar[1].pos,
			}
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.bin = &BoolLit{
				Pos:  yyDollar[1].pos,
				Bool: true,
			}
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.bin = &BoolLit{
				Pos:  yyDollar[1].pos,
				Bool: false,
			}
		}
	case 66:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.cas = []*Case{yyDollar[1].ca}
		}
	case 67:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.cas = append(yyDollar[1].cas, yyDollar[2].ca)
		}
	case 68:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.ca = &Case{
				Name: yyDollar[2].id,
				Type: yyDollar[4].id,
				Body: yyDollar[6].exp,
			}
		}
	case 69:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.ca = &Case{
				Name: &Ident{
					Name: "null",
					Pos:  yyDollar[2].pos,
				},
				Type: &Ident{
					Name: "Null",
					Pos:  yyDollar[2].pos,
				},
				Body: yyDollar[4].exp,
			}
		}
	}
	goto yystack /* stack new state and value */
}
