package structures

import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](ls ...T) T {
	if len(ls) == 0 {
		panic("Error: Max run on 0 elements")
	}
	currentMax := ls[0]
	for _, v := range ls[1:] {
		if v > currentMax {
			currentMax = v
		}
	}
	return currentMax
}

func Min[T constraints.Ordered](ls ...T) T {
	if len(ls) == 0 {
		panic("Error: Max run on 0 elements")
	}
	currentMin := ls[0]
	for _, v := range ls[1:] {
		if v < currentMin {
			currentMin = v
		}
	}
	return currentMin
}

// FIXME: This should only be for testing. Remove when we can.
func PANIC_ON_ERR(err error) {
	if err != nil {
		panic(err)
	}
}
