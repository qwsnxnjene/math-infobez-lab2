package main

// bentFunc строит bent-функцию от n переменных (n чётное):
// f = x0*x(n/2) + x1*x(n/2+1) + ... + x(n/2-1)*x(n-1)
//
// n=4: f = x0x2 + x1x3
// n=6: f = x0x3 + x1x4 + x2x5
//
// Нелинейность такой функции равна 2^(n-1) - 2^(n/2-1), это максимум.
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

// bentMaxNL - теоретическое значение максимальной нелинейности (таблица 4)
func bentMaxNL(n int) int {
	return (1 << (n - 1)) - (1 << (n/2 - 1))
}
