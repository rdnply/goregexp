package stack

type Stack []interface{}

func New() Stack {
	return make([]interface{}, 0)
}

func (s *Stack) IsEmpty() bool {
	return len(*s) == 0
}

func (s *Stack) Push(el interface{}) {
	*s = append(*s, el)
}

func (s *Stack) Pop() (interface{}, bool) {
	if s.IsEmpty() {
		return ' ', false
	}

	index := len(*s) - 1
	top := (*s)[index]
	*s = (*s)[:index]

	return top, true
}

func (s *Stack) Top() interface{} {
	if s.IsEmpty() {
		return ' '
	}

	return (*s)[len(*s)-1]
}
