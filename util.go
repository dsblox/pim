package main

func findInt(s []int, valToFind int) int {
	for i, v := range s {
		if (v == valToFind) {
			return i
		}
	}
	return -1
}

func removeIntByIndex(s []int, idxToRemove int) []int {
	return append(s[:idxToRemove], s[idxToRemove+1:]...)
}

func removeIntByValue(s []int, valToRemove int) []int {
	idxToRemove := findInt(s, valToRemove)
	if idxToRemove != -1 {
		return removeIntByIndex(s, idxToRemove)
	}
	return s
}