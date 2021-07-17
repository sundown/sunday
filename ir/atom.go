package ir

import (
	"fmt"
	"sundown/sunday/parser"
)

type Atom struct {
	TypeOf *Type
	Tuple  []*Expression
	Vector []*Expression
	Int    *int64
	Nat    *uint64
	Real   *float64
	Bool   *bool
	Str    *string
	Noun   *Ident
	Param  *uint
}

func (a *Atom) String() string {
	switch {
	case a.Int != nil:
		return fmt.Sprint(*a.Int)
	case a.Nat != nil:
		return fmt.Sprint(*a.Nat)
	case a.Real != nil:
		return fmt.Sprint(*a.Real)
	case a.Bool != nil:
		if *a.Bool {
			return "True"
		} else {
			return "False"
		}
	case a.Str != nil:
		return *a.Str
	case a.Noun != nil:
		return *a.Noun.Namespace + "::" + *a.Noun.Ident
	case a.Param != nil:
		return "@"
	case a.Vector != nil:
		var str string
		for _, expr := range a.Vector {
			str += ", " + expr.String()
		}

		return "[" + str[2:] + "]"
	case a.Tuple != nil:
		var str string
		for _, expr := range a.Tuple {
			str += ", " + expr.String()
		}

		return "(" + str[2:] + ")"
	}

	return "_"
}

func (t *Type) AsVector() *Type {
	return &Type{Vector: t}
}

func (state *State) AnalyseAtom(primary *parser.Primary) (a *Atom) {
	switch {
	case primary.Tuple != nil:
		var types []*Type
		var strct []*Expression
		for _, expr := range primary.Tuple {
			e := state.AnalyseExpression(expr)
			types = append(types, e.TypeOf)
			strct = append(strct, e)
		}

		a = &Atom{TypeOf: &Type{Tuple: types}, Tuple: strct}
	case primary.Vec != nil:
		var vec []*Expression
		for _, expr := range primary.Vec {
			e := state.AnalyseExpression(expr)
			/* all elements must be of same type */
			// Can't compare types properly yet
			/* if index > 0 && vec[index-1].TypeOf != e.TypeOf {
				panic("ir: Atom: Vector: divergent type at position: " + fmt.Sprint(index) + "\n" + e.TypeOf.String() + " & " + vec[index-1].TypeOf.String())
			} */

			vec = append(vec, e)
		}

		a = &Atom{TypeOf: vec[0].TypeOf, Vector: vec}
	case primary.Int != nil:
		a = &Atom{TypeOf: BaseType("Int"), Int: primary.Int}
	case primary.Real != nil:
		a = &Atom{TypeOf: BaseType("Real"), Real: primary.Real}
	case primary.Bool != nil:
		/* TODO: add a third bool state "Maybe", maybe */
		var b bool
		if *primary.Bool == "True" {
			b = true
		} else {
			b = false
		}

		a = &Atom{TypeOf: BaseType("Bool"), Bool: &b}
	case primary.String != nil:
		/* TODO: strings might need their "" cut off each end because parser sometimes leaves them */
		a = &Atom{TypeOf: BaseType("String"), Str: primary.String}
	case primary.Noun != nil:
		a = state.GetNoun(IRIdent(primary.Noun))
	case primary.Param != nil:
		/* TODO: add param index if it exists, needs parser modification too */
		a = &Atom{TypeOf: AtomicType("Param")} /* Currently dead */
	}

	return a
}
