package main

import (
	"fmt"
	"unicode"
)

type (
	Acceptor func(rune) bool
	state    func(rune) (bool, state) // bool in return expr must eval before state
)

func Acc(spec interface{}) (acc Acceptor) {
	switch spec := spec.(type) {
	case Acceptor:
		return spec
	case rune:
		return func(rn rune) bool {
			return rn == spec
		}
	case []rune:
		return func(rn rune) bool {
			for _, ex := range spec {
				if rn == ex {
					return true
				}
			}
			return false
		}
	case string:
		return func(rn rune) bool {
			for _, ex := range spec {
				if rn == ex {
					return true
				}
			}
			return false
		}
	case *unicode.RangeTable:
		return func(rn rune) bool {
			return unicode.Is(spec, rn)
		}
	case []*unicode.RangeTable:
		return func(rn rune) bool {
			return unicode.In(rn, spec...)
		}
	default:
		return nil
	}
}

func All() Acceptor {
	return func(rn rune) bool {
		return true
	}
}

func None() Acceptor {
	return func(rn rune) bool {
		return false
	}
}

func statify(acc Acceptor) state {
	var internal state
	internal = func(rn rune) (bool, state) {
		return acc(rn), internal
	}
	return internal
}

/********** Acceptor Composition Functions **********/

func Branch(mapping map[rune]Acceptor, alt Acceptor) Acceptor {
	var internal state = func(rn rune) (bool, state) {
		if next, ok := mapping[rn]; ok {
			return true, statify(next)
		}
		return alt(rn), statify(alt)
	}
	return func(rn rune) (out bool) {
		out, internal = internal(rn)
		return
	}
}

func Skip(skip int, acc Acceptor) Acceptor {
	return Chain(Truncate(skip, All()), acc)
}

func FirstOf(accs ...Acceptor) Acceptor {
	var internal state = func(rn rune) (bool, state) {
		for _, acc := range accs {
			if acc(rn) {
				return true, statify(acc)
			}
		}
		return false, nil
	}
	return func(rn rune) (out bool) {
		out, internal = internal(rn)
		return out
	}
}

func Truncate(max int, acc Acceptor) Acceptor {
	count := 0
	return func(rn rune) bool {
		if count < max && acc(rn) {
			count++
			return true
		}
		return false
	}
}

func EndsBefore(delim rune, acc Acceptor) Acceptor {
	return func(rn rune) bool {
		if rn == delim {
			return false
		}
		return acc(rn)
	}
}

func EndsWith(delim rune, acc Acceptor) Acceptor {
	seen := false
	return func(rn rune) bool {
		if seen {
			return false
		}
		if rn == delim {
			seen = true
			return true
		}
		return acc(rn)
	}
}

func Chain(accs ...Acceptor) Acceptor {
	i, n := 0, len(accs)
	return func(rn rune) bool {
		for ; i != n; i++ {
			if accs[i](rn) {
				return true
			}
		}
		return false
	}
}

/********** Acceptor Assertions **********/

func AssertStart(acc Acceptor, err error) Acceptor {
	var internal state = func(rn rune) (bool, state) {
		if !acc(rn) {
			panic(err)
		}
		return true, statify(acc)
	}
	return func(rn rune) (out bool) {
		out, internal = internal(rn)
		return
	}
}

func AssertAtMost(acc Acceptor, most int, err error) Acceptor {
	count := 0
	return func(rn rune) bool {
		if acc(rn) {
			if count == most {
				panic(err)
			}
			count++
			return true
		}
		return false
	}
}

/********** Demo **********/

func main() {
	acceptor    :=  EndsBefore('m', All())
	acceptorSrc := `EndsBefore('m', All())`

	src := "This is a demo."

	fmt.Printf("\n")
	fmt.Printf("Acceptor: \"%s\"\n\n", acceptorSrc)
	fmt.Printf("Source:   \"%s\"\n\n", src)

	for i, rn := range src {
		if !acceptor(rn) {
			fmt.Printf("--------\n")
			fmt.Printf("Rejected '%c' at %d ...\n\n", rn, i)
			fmt.Printf("Consumed:  \"%s\"\n\n", src[:i+1])
			fmt.Printf("Remaining: \"%s\"\n\n", src[i+1:])
			break
		}
		fmt.Printf("Accepted '%c' at %d ...\n", rn, i)
	}

	fmt.Printf("Program complete.\n\n")
}

