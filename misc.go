package main

//type Orderable interface {
//	index() constraints.Ordered
//}
//
//func ordonne(unsorted []Orderable) {
//	sort.Slice(unsorted, func(i, j int) bool {
//		return unsorted[i].index() < unsorted[j].index()
//	})
//}

func mapSlice[T any, M any](a []T, f func(T) M) []M {
	n := make([]M, len(a))
	for i, e := range a {
		n[i] = f(e)
	}
	return n
}

func mapifySlice[E any, I comparable](a []E, f func(E) I) map[I]E {
	n := make(map[I]E)
	for _, e := range a {
		n[f(e)] = e
	}
	return n
}

func mapSelect[E any, I comparable](m map[I]E, f func(I, E) bool) map[I]E {
	n := make(map[I]E)
	for i, e := range m {
		if f(i, e) {
			n[i] = e
		}
	}
	return n
}

func contains[E comparable](array []E, item E) bool {
	if array == nil {
		return false
	}
	for _, s := range array {
		if item == s {
			return true
		}
	}
	return false
}

func intersect[E comparable](elementsA []E, elementsB []E) (both []E, onlyA []E, onlyB []E) {
	for _, elementA := range elementsA {
		foundBoth := false
		for _, elementB := range elementsB {
			if elementA == elementB {
				both = append(both, elementA)
				foundBoth = true
			}
		}
		if !foundBoth {
			onlyA = append(onlyA, elementA)
		}
	}

	for _, elementB := range elementsB {
		foundBoth := false
		for _, elementA := range elementsA {
			if elementA == elementB {
				foundBoth = true
			}
		}
		if !foundBoth {
			onlyB = append(onlyB, elementB)
		}
	}
	return both, onlyA, onlyB
}
