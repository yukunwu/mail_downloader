package main

type ExtractFunc[T any] func(int) T

func extract[T any](len int, f ExtractFunc[T]) []T {
	var res []T
	for i := 0; i < len; i++ {
		res = append(res, f(i))
	}
	return res
}
func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}
