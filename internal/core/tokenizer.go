package core

// TokenKind represents the type of a Brainfuck instruction token.
type TokenKind int

// Token kinds for each Brainfuck instruction.
const (
	TokInvalid    TokenKind = iota // invalid/unknown token
	TokShiftRight                  // > : move pointer right
	TokShiftLeft                   // < : move pointer left
	TokAdd                         // + : increment cell
	TokSub                         // - : decrement cell
	TokOut                         // . : output cell
	TokIn                          // , : input to cell
	TokLBracket                    // [ : begin loop
	TokRBracket                    // ] : end loop
	TokEOF                         // end of file marker
)

// tokenNames maps each TokenKind to its string representation for debugging.
var tokenNames = [...]string{
	TokInvalid:    "TokInvalid",
	TokShiftRight: "TokShiftRight",
	TokShiftLeft:  "TokShiftLeft",
	TokAdd:        "TokAdd",
	TokSub:        "TokSub",
	TokOut:        "TokOut",
	TokIn:         "TokIn",
	TokLBracket:   "TokLBracket",
	TokRBracket:   "TokRBracket",
	TokEOF:        "TokEOF",
}

// String returns the string representation of the TokenKind.
func (k TokenKind) String() string {
	return tokenNames[k]
}

// Token represents a single lexical token from the source.
type Token struct {
	Kind TokenKind // the type of token
	Pos  Position  // location in source
}

// charToToken maps Brainfuck command characters to their token kinds.
var charToToken = [...]TokenKind{
	'>': TokShiftRight,
	'<': TokShiftLeft,
	'+': TokAdd,
	'-': TokSub,
	'.': TokOut,
	',': TokIn,
	'[': TokLBracket,
	']': TokRBracket,
}

// FoldToken counts consecutive tokens of the given kind starting at index i.
// Returns the count of matching tokens found. If the token at index i doesn't
// match the given kind, returns 0.
func FoldToken(tokens []Token, i int, kind TokenKind) int {
	count := 0
	for ; i < len(tokens) && tokens[i].Kind == kind; i++ {
		count++
	}
	return count
}

// Tokenize converts Brainfuck source code into a slice of tokens.
// Non-command characters are ignored. The returned slice always ends
// with a TokEOF token.
func Tokenize(src []byte) []Token {
	// Setting capacity slightly smaller for whitespace
	tokens := make([]Token, 0, len(src)/2)

	line, col := 1, 1
	for i, b := range src {
		if kind := charToToken[b]; kind != 0 {
			tokens = append(tokens, Token{
				Kind: kind,
				Pos:  Position{Offset: i, Line: line, Column: col},
			})
		} else if b == '\n' {
			line++
			col = 0
		}
		col++
	}

	// Add the EOF token
	tokens = append(tokens, Token{
		Kind: TokEOF,
		Pos:  Position{Offset: len(src), Line: line, Column: col},
	})

	return tokens
}
