%{
//go:generate go tool yacc -l syntax.y

package ast

import "go/token"

func init() {
	yyErrorVerbose = true
}
%}

%union {
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

%type<cls> classes
%type<cl>  class
%type<fts> feature_list
%type<ft>  feature var method
%type<fms> var_formals var_formals_nonempty formals formals_nonempty
%type<fm>  var_formal formal
%type<ext> extends
%type<exp> block block_nonempty expr primary
%type<act> actuals actuals_nonempty
%type<cas> cases
%type<ca>  case
%type<bin> boolean

%token<pos> CLASS EXTENDS NATIVE VAR DEF OVERRIDE SUPER NEW ELSE NULL THIS ARROW CASE TRUE FALSE ILLEGAL '('
%token<id>  TYPEID OBJECTID
%token<int> INTEGER
%token<str> STRING

%left<pos> '='
%left<pos> IF WHILE
%left<pos> MATCH
%left<pos> LE '<'
%left<pos> EQ
%left<pos> '+' '-'
%left<pos> '*' '/'
%left<pos> '!'
%left<pos> '.'

%%

program
: classes
	{
		p := yylex.(*lex).program
		p.Classes = append(p.Classes, $1...)
	}
;

classes
: /* empty */
	{
		$$ = nil
	}
| classes class
	{
		$$ = append($1, $2)
	}
;

class
: CLASS TYPEID '(' var_formals ')' extends '{' feature_list '}'
	{
		$$ = &Class{
			Type:     $2,
			Formals:  $4,
			Extends:  $6,
			Features: $8,
		}
	}
;

extends
: /* empty */
	{
		$$ = &Extends{
			Type: &Ident{
				Name: "Any",
				Pos:  0,
			},
			Args: nil,
		}
	}
| EXTENDS TYPEID '(' actuals ')'
	{
		$$ = &Extends{
			Type: $2,
			Args: $4,
		}
	}
| EXTENDS NATIVE
	{
		$$ = &Extends{
			Type: &Ident{
				Name: "native",
				Pos:  $2,
			},
			Args: nil,
		}
	}
;

feature_list
: /* empty */
	{
		$$ = nil
	}
| feature_list feature ';'
	{
		$$ = append($1, $2)
	}
;

feature
: '{' block '}'
	{
		$$ = &Init{
			Expr: $2,
		}
	}
| var
	{
		$$ = $1
	}
| method
	{
		$$ = $1
	}
| OVERRIDE method
	{
		$$ = $2
		$$.(*Method).Override = true
	}
;

var
: VAR OBJECTID ':' TYPEID '=' expr
	{
		$$ = &Attribute{
			Name: $2,
			Type: $4,
			Init: $6,
		}
	}
| VAR OBJECTID '=' NATIVE
	{
		$$ = &Attribute{
			Name: $2,
			Type: &Ident{
				Name: "native",
				Pos:  $4,
			},
			Init: &NativeExpr{
				Pos: $4,
			},
		}
	}
;

method
: DEF OBJECTID '(' formals ')' ':' TYPEID '=' expr
	{
		$$ = &Method{
			Name: $2,
			Args: $4,
			Type: $7,
			Body: $9,
		}
	}
| DEF OBJECTID '(' formals ')' ':' TYPEID '=' NATIVE
	{
		$$ = &Method{
			Name: $2,
			Args: $4,
			Type: $7,
			Body: &NativeExpr{
				Pos: $9,
			},
		}
	}
;

actuals
: /* empty */
	{
		$$ = nil
	}
| actuals_nonempty
	{
		$$ = $1
	}
;

actuals_nonempty
: expr
	{
		$$ = []Expr{$1}
	}
| actuals_nonempty ',' expr
	{
		$$ = append($1, $3)
	}
;

block
: /* empty */
	{
		$$ = &UnitExpr{
			Pos: yylex.(*lex).Pos(),
		}
	}
| block_nonempty
	{
		$$ = $1
	}
;

block_nonempty
: expr
	{
		$$ = $1
	}
| VAR OBJECTID ':' TYPEID '=' expr ';' block_nonempty
	{
		$$ = &VarExpr{
			Name: $2,
			Type: $4,
			Init: $6,
			Body: $8,
		}
	}
| expr ';' block_nonempty
	{
		$$ = &ChainExpr{
			Pre:  $1,
			Expr: $3,
		}
	}
;

var_formals
: /* empty */
	{
		$$ = nil
	}
| var_formals_nonempty
	{
		$$ = $1
	}
;

var_formals_nonempty
: var_formal
	{
		$$ = []*Formal{$1}
	}
| var_formals_nonempty ',' var_formal
	{
		$$ = append($1, $3)
	}
;

var_formal
: VAR formal
	{
		$$ = $2
	}
;

formals
: /* empty */
	{
		$$ = nil
	}
| formals_nonempty
	{
		$$ = $1
	}
;

formals_nonempty
: formal
	{
		$$ = []*Formal{$1}
	}
| formals_nonempty ',' formal
	{
		$$ = append($1, $3)
	}
;

formal
: OBJECTID ':' TYPEID
	{
		$$ = &Formal{
			Name: $1,
			Type: $3,
		}
	}
;

expr
: primary
	{
		$$ = $1
	}
| OBJECTID '=' expr %prec '='
	{
		$$ = &AssignExpr{
			Name: $1,
			Expr: $3,
		}
	}
| '!' expr %prec '!'
	{
		$$ = &NotExpr{
			Expr: $2,
		}
	}
| '-' expr %prec '!'
	{
		$$ = &NegativeExpr{
			Expr: $2,
		}
	}
| IF '(' expr ')' expr ELSE expr %prec IF
	{
		$$ = &IfExpr{
			Cond: $3,
			Then: $5,
			Else: $7,
		}
	}
| WHILE '(' expr ')' expr %prec WHILE
	{
		$$ = &WhileExpr{
			Cond: $3,
			Body: $5,
		}
	}
| expr LE expr %prec LE
	{
		$$ = &LessOrEqualExpr{
			Pos:     $2,
			Left:    $1,
			Right:   $3,
		}
	}
| expr '<' expr %prec '<'
	{
		$$ = &LessThanExpr{
			Pos:     $2,
			Left:    $1,
			Right:   $3,
		}
	}
| expr EQ expr %prec EQ
	{
		$$ = &DynamicCallExpr{
			Recv: $1,
			Name: &Ident{
				Name: "equals",
				Pos:  $2,
			},
			Args: []Expr{
				$3,
			},
		}
	}
| expr '*' expr %prec '*'
	{
		$$ = &MultiplyExpr{
			Pos:   $2,
			Left:  $1,
			Right: $3,
		}
	}
| expr '/' expr %prec '/'
	{
		$$ = &DivideExpr{
			Pos:   $2,
			Left:  $1,
			Right: $3,
		}
	}
| expr '+' expr %prec '+'
	{
		$$ = &AddExpr{
			Pos:   $2,
			Left:  $1,
			Right: $3,
		}
	}
| expr '-' expr %prec '-'
	{
		$$ = &SubtractExpr{
			Pos:   $2,
			Left:  $1,
			Right: $3,
		}
	}
| expr MATCH '{' cases '}' %prec MATCH
	{
		$$ = &MatchExpr{
			Pos:   $2,
			Left:  $1,
			Cases: $4,
		}
	}
| expr '.' OBJECTID '(' actuals ')' %prec '.'
	{
		$$ = &DynamicCallExpr{
			Recv: $1,
			Name: $3,
			Args: $5,
		}
	}
;

primary
: OBJECTID '(' actuals ')'
	{
		$$ = &DynamicCallExpr{
			Recv: &ThisExpr{
				Pos: $1.Pos,
			},
			Name: $1,
			Args: $3,
		}
	}
| SUPER '.' OBJECTID '(' actuals ')'
	{
		$$ = &SuperCallExpr{
			Pos:  $1,
			Name: $3,
			Args: $5,
		}
	}
| NEW TYPEID '(' actuals ')'
	{
		$$ = &StaticCallExpr{
			Recv: &AllocExpr{
				Type: $2,
			},
			// $2 refers to a type, so we need to make a copy
			// because this one is the constructor method:
			Name: &Ident{
				Name: $2.Name,
				Pos:  $2.Pos,
			},
			Args: $4,
			
		}
	}
| '{' block '}'
	{
		$$ = $2
	}
| '(' expr ')'
	{
		$$ = $2
	}
| NULL
	{
		$$ = &NullExpr{
			Pos: $1,
		}
	}
| '(' ')'
	{
		$$ = &UnitExpr{
			Pos: $1,
		}
	}
| OBJECTID
	{
		$$ = &NameExpr{
			Name: $1,
		}
	}
| INTEGER
	{
		$$ = &IntExpr{
			Lit: $1,
		}
	}
| STRING
	{
		$$ = &StringExpr{
			Lit: $1,
		}
	}
| boolean
	{
		$$ = &BoolExpr{
			Lit: $1,
		}
	}
| THIS
	{
		$$ = &ThisExpr{
			Pos: $1,
		}
	}
;

boolean
: TRUE
	{
		$$ = &BoolLit{
			Pos:  $1,
			Bool: true,
		}
	}
| FALSE
	{
		$$ = &BoolLit{
			Pos:  $1,
			Bool: false,
		}
	}
;

cases
: case
	{
		$$ = []*Case{$1}
	}
| cases case
	{
		$$ = append($1, $2)
	}
;

case
: CASE OBJECTID ':' TYPEID ARROW block
	{
		$$ = &Case{
			Name: $2,
			Type: $4,
			Body: $6,
		}
	}
| CASE NULL ARROW block
	{
		$$ = &Case{
			Name: &Ident{
				Name: "null",
				Pos:  $2,
			},
			Type: &Ident{
				Name: "Null",
				Pos:  $2,
			},
			Body: $4,
		}
	}
;
