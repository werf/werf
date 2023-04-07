package util

import "fmt"

type Pair[T1, T2 any] struct {
	First  T1
	Second T2
}

func NewPair[T1, T2 any](first T1, second T2) Pair[T1, T2] {
	return Pair[T1, T2]{First: first, Second: second}
}

func (p Pair[T1, T2]) Unpair() (T1, T2) {
	return p.First, p.Second
}

func (p Pair[T1, T2]) String() string {
	return fmt.Sprintf("[%v %v]", p.First, p.Second)
}
