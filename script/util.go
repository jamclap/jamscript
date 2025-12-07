package script

type Stack[T any] []T

func last[T any](items *[]T) *T {
	return &(*items)[len(*items)-1]
}

func (s *Stack[T]) Pop() T {
	i := len(*s) - 1
	v := (*s)[i]
	*s = (*s)[:i]
	return v
}

func pop[T any](items *[]T) T {
	return (*Stack[T])(items).Pop()
}

func (s *Stack[T]) Push(v T) *T {
	*s = append(*s, v)
	return last((*[]T)(s))
}

func push[T any](items *[]T, item T) *T {
	return (*Stack[T])(items).Push(item)
}
