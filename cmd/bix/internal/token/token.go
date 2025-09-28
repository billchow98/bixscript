package token

type Type int

const (
	Error Type = iota
	Eof
	Newline
	Plus
	PlusEqual
	Minus
	MinusEqual
	Star
	StarEqual
	StarStar
	StarStarEqual
	Slash
	SlashEqual
	Bang
	BangEqual
	Equal
	EqualEqual
	Less
	LessEqual
	Greater
	GreaterEqual
	Dot
	LeftParen
	RightParen
	Identifier
	Let
	True
	False
	Number
	String
	LeftBrace
	RightBrace
	Semicolon
	Comma
	If
	Else
	And
	Or
	While
	For
	Function
	Return
)

type Token struct {
	Type  Type
	Value string
	File  string
	Line  int
	Col   int
}
