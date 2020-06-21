package goregexp

import (
	"fmt"
	"github.com/rdnply/goregexp/stack"
	"strings"
)

func insertExplicitConcatOp(regexp string) string {
	const concatOp = '.'
	var res strings.Builder

	for i, c := range regexp {
		res.WriteRune(c)

		if c == '(' || c == '|' {
			continue
		}

		if i < len(regexp)-1 {
			next := []rune(regexp)[i+1]

			if next == '*' || next == '+' || next == '?' || next == '|' || next == ')' {
				continue
			}

			res.WriteRune(concatOp)
		}
	}

	return res.String()
}

func getPriority(ch rune) int {
	switch ch {
	case '*', '+', '?':
		return 3
	case '.':
		return 2
	case '|':
		return 1
	default:
		return -1
	}
}

func getTopRune(st stack.Stack) rune {
	t := st.Top()
	top, _ := t.(rune)
	return top
}

func isOperation(c rune) bool {
	return c == '*' || c == '|' || c == '.' || c == '+' || c == '?'
}

func toPostfix(regexp string) (string, error) {
	var postfix strings.Builder
	st := stack.New()

	for _, ch := range regexp {
		switch {
		case ch >= 'a' && ch <= 'z':
			postfix.WriteRune(ch)
		case ch == '(':
			st.Push(ch)
		case ch == ')':
			for !st.IsEmpty() && st.Top() != '(' {
				t, _ := st.Pop()
				top, _ := t.(rune)
				postfix.WriteRune(top)
			}

			if st.Top() != '(' {
				return "", fmt.Errorf("missing open bracket in expression")
			}

			st.Pop()
		default:
			for !st.IsEmpty() && getPriority(ch) <= getPriority(getTopRune(st)) {
				t, _ := st.Pop()
				top, _ := t.(rune)
				postfix.WriteRune(top)
			}

			st.Push(ch)
		}
	}

	for !st.IsEmpty() {
		t, _ := st.Pop()
		top, _ := t.(rune)
		if !isOperation(top) {
			return "", fmt.Errorf("missing close bracket in expression")
		}

		postfix.WriteRune(top)
	}

	return postfix.String(), nil
}

type state struct {
	isEnd           bool
	epsTransitions  []*state
	symbTransitions map[rune]*state
}

type expr struct {
	start, end *state
}

func createState(isEnd bool) *state {
	return &state{
		isEnd:           isEnd,
		epsTransitions:  make([]*state, 0),
		symbTransitions: make(map[rune]*state),
	}
}

func (st *state) addEpsTransition(to *state) {
	st.epsTransitions = append(st.epsTransitions, to)
}

func (st *state) addSymbTransition(to *state, s rune) {
	st.symbTransitions[s] = to
}

func fromSymbol(s rune) *expr {
	start := createState(false)
	end := createState(true)
	start.addSymbTransition(end, s)

	return &expr{start, end}
}

func fromEps() *expr {
	start := createState(false)
	end := createState(true)
	start.addEpsTransition(end)

	return &expr{start, end}
}

func concat(fi *expr, se *expr) *expr {
	fi.end.addEpsTransition(se.start)
	fi.end.isEnd = false

	return &expr{fi.start, se.end}
}

func union(fi *expr, se *expr) *expr {
	start := createState(false)
	end := createState(true)

	start.addEpsTransition(fi.start)
	start.addEpsTransition(se.start)
	fi.end.addEpsTransition(end)
	se.end.addEpsTransition(end)

	fi.end.isEnd = false
	se.end.isEnd = false

	return &expr{start, end}
}

func closure(nfa *expr) *expr {
	start := createState(false)
	end := createState(true)

	start.addEpsTransition(end)
	start.addEpsTransition(nfa.start)
	nfa.end.addEpsTransition(end)
	nfa.end.addEpsTransition(nfa.start)

	nfa.end.isEnd = false

	return &expr{start, end}
}

func plus(nfa *expr) *expr {
	start := createState(false)
	end := createState(true)

	start.addEpsTransition(nfa.start)
	nfa.end.addEpsTransition(end)
	nfa.end.addEpsTransition(nfa.start)

	nfa.end.isEnd = false

	return &expr{start, end}
}

func questionMark(nfa *expr) *expr {
	start := createState(false)
	end := createState(true)

	start.addEpsTransition(end)
	start.addEpsTransition(nfa.start)
	nfa.end.addEpsTransition(end)

	nfa.end.isEnd = false

	return &expr{start, end}
}

func pairFromTop(st *stack.Stack) (*expr, *expr) {
	r, _ := st.Pop()
	l, _ := st.Pop()
	right, _ := r.(*expr)
	left, _ := l.(*expr)

	return left, right
}

func NFAFromTop(st *stack.Stack) *expr {
	t, _ := st.Pop()
	top, _ := t.(*expr)

	return top
}

func toNFA(postfix string) *expr {
	if len(postfix) == 0 {
		return fromEps()
	}

	st := stack.New()

	for _, c := range postfix {
		switch c {
		case '.':
			st.Push(concat(pairFromTop(&st)))
		case '|':
			st.Push(union(pairFromTop(&st)))
		case '*':
			st.Push(closure(NFAFromTop(&st)))
		case '+':
			st.Push(plus(NFAFromTop(&st)))
		case '?':
			st.Push(questionMark(NFAFromTop(&st)))
		default:
			st.Push(fromSymbol(c))
		}
	}

	return NFAFromTop(&st)
}

func getNextStates(from *state) []*state {
	visited := make(map[*state]bool)

	return getStates(from, visited)
}

func getStates(from *state, visited map[*state]bool) []*state {
	if len(from.epsTransitions) == 0 {
		return []*state{from}
	}

	next := make([]*state, 0)
	for _, t := range from.epsTransitions {
		if !visited[t] {
			visited[t] = true
			next = append(next, getStates(t, visited)...)
		}
	}

	return next
}

func containsTerminal(states []*state) bool {
	for _, s := range states {
		if s.isEnd {
			return true
		}
	}

	return false
}

func search(nfa *expr, word string) bool {
	curStates := getNextStates(nfa.start)

	for _, c := range word {
		nextStates := make([]*state, 0)
		for _, curSt := range curStates {
			if nextState, ok := curSt.symbTransitions[c]; ok {
				nextStates = append(nextStates, getNextStates(nextState)...)
			}
		}

		curStates = nextStates
	}

	return containsTerminal(curStates)
}

func CreateMatcher(regexp string) (func(string) bool, error) {
	postfix, err := toPostfix(insertExplicitConcatOp(regexp))
	if err != nil {
		return nil, err
	}

	nfa := toNFA(postfix)

	return func(word string) bool {
		return search(nfa, word)
	}, nil
}
