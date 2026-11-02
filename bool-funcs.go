package main

// Булева функция от n переменных хранится как срез из 2^n нулей и единиц
// (это её таблица истинности).
// Индекс в срезе - номер аргумента, записанный битами x0 x1 ... x(n-1),
// где x0 - СТАРШИЙ бит. Например, для n=3 индекс 5 = 101 это x0=1, x1=0, x2=1.

// numVars считает, сколько у функции переменных (длина среза = 2^n)
func numVars(f []int) int {
	n := 0
	for (1 << n) < len(f) {
		n++
	}
	return n
}

// hammingWeight - норма Хэмминга
func hammingWeight(x int) int {
	w := 0
	for x > 0 {
		w += x & 1
		x >>= 1
	}
	return w
}

// parity - чётность числа единичных бит.
func parity(x int) int {
	return hammingWeight(x) % 2
}

// distance - расстояние Хэмминга между двумя функциями (или векторами)
func distance(f, g []int) int {
	d := 0
	for i := range f {
		if f[i] != g[i] {
			d++
		}
	}
	return d
}

// xorFunc - сумма двух функций по модулю 2
func xorFunc(f, g []int) []int {
	res := make([]int, len(f))
	for i := range f {
		res[i] = f[i] ^ g[i]
	}
	return res
}

// addOne - функция f + 1
func addOne(f []int) []int {
	res := make([]int, len(f))
	for i := range f {
		res[i] = f[i] ^ 1
	}
	return res
}

// makeFunc строит таблицу истинности по обычной go-функции.
// В неё передаётся срез x из n битов (x[0] - старший).
func makeFunc(n int, fn func(x []int) int) []int {
	f := make([]int, 1<<n)
	for idx := 0; idx < len(f); idx++ {
		x := make([]int, n)
		for i := 0; i < n; i++ {
			x[i] = (idx >> (n - 1 - i)) & 1
		}
		f[idx] = fn(x)
	}
	return f
}

// affine - линейная функция c + a*x.
func affine(n, a, c int) []int {
	f := make([]int, 1<<n)
	for x := 0; x < len(f); x++ {
		f[x] = parity(a&x) ^ c
	}
	return f
}

// nonlinearity - нелинейность функции: минимальное расстояние
// до линейных функций. Перебираем все 2^(n+1) штук.
func nonlinearity(f []int) int {
	n := numVars(f)
	best := len(f) // больше расстояния быть не может
	for a := 0; a < (1 << n); a++ {
		for c := 0; c < 2; c++ {
			d := distance(f, affine(n, a, c))
			if d < best {
				best = d
			}
		}
	}
	return best
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
