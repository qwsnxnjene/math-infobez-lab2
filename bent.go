package main

// bentFunc строит bent-функцию от n переменных
func bentFunc(n int) []int {
	half := n / 2
	return makeFunc(n, func(x []int) int {
		sum := 0
		for i := 0; i < half; i++ {
			sum ^= x[i] & x[i+half]
		}
		return sum
	})
}

// bentMaxNL - значение максимальной нелинейности
func bentMaxNL(n int) int {
	return (1 << (n - 1)) - (1 << (n/2 - 1))
}
