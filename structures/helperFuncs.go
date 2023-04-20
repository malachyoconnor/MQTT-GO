// Package structures contains helper functions and structs for other packages.
// This includes functions for finding the max and min of a list of numbers,
// linked lists, printing functions, a thread safe map implementation and
// a ticket implementation.
package structures

import "golang.org/x/exp/constraints"

// Max returns the maximum value in a list of comparable values.
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

// Min returns the minimum value in a list of comparable values.
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
