package main

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

func selectMap[E any, I comparable](m map[I]E, filter func(I, E) bool) map[I]E {
	n := make(map[I]E)
	for i, e := range m {
		if filter(i, e) {
			n[i] = e
		}
	}
	return n
}

func selectMapByValue[Key comparable, Value any](m map[Key]Value, selector func(Value) bool) map[Key]Value {
	return selectMap(m, func(_ Key, value Value) bool {
		return selector(value)
	})
}

func selectSlice[Element any](slice []Element, test func(Element) bool) []Element {
	var newSlice []Element
	for _, element := range slice {
		if test(element) {
			newSlice = append(newSlice, element)
		}
	}
	return newSlice
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

//func containsFunc[E comparable](array []E) func(item E) bool {
//	return func(item E) bool {
//		return contains(array, item)
//	}
//}

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

func keys[Key comparable, Element any](m map[Key]Element) []Key {
	var ks []Key
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func values[Key comparable, Element any](m map[Key]Element) []Element {
	var es []Element
	for _, e := range m {
		es = append(es, e)
	}
	return es
}
