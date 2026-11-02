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

// hammingWeight - норма Хэмминга числа: сколько в нём единичных бит
func hammingWeight(x int) int {
	w := 0
	for x > 0 {
		w += x & 1
		x >>= 1
	}
	return w
}

// parity - чётность числа единичных бит.
// Это же скалярное произведение a*x по модулю 2, если подать a & x
func parity(x int) int {
	return hammingWeight(x) % 2
}

// distance - расстояние Хэмминга между двумя функциями (или векторами):
// сколько раз значения в таблицах истинности различаются
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

// addOne - функция f + 1 (все значения инвертируются)
func addOne(f []int) []int {
	res := make([]int, len(f))
	for i := range f {
		res[i] = f[i] ^ 1
	}
	return res
}

// makeFunc строит таблицу истинности по обычной функции.
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

// affine - аффинная функция c + a*x.
// a - коэффициенты при x0..x(n-1), записанные как биты числа (как и x),
// c - свободный член (0 или 1)
func affine(n, a, c int) []int {
	f := make([]int, 1<<n)
	for x := 0; x < len(f); x++ {
		f[x] = parity(a&x) ^ c
	}
	return f
}

// nonlinearity - нелинейность функции: минимальное расстояние
// до аффинных функций. Перебираем все 2^(n+1) штук.
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

// walsh - преобразование Уолша-Адамара.
// w[a] = сумма по x от (-1)^(f(x) + a*x).
// Чем больше |w[a]|, тем ближе f к аффинной функции с коэффициентами a.
func walsh(f []int) []int {
	w := make([]int, len(f))
	for i := range f {
		w[i] = 1 - 2*f[i] // 0 -> 1, 1 -> -1
	}
	// "бабочки", как в быстром преобразовании Фурье
	for step := 1; step < len(w); step *= 2 {
		for i := 0; i < len(w); i += 2 * step {
			for j := i; j < i+step; j++ {
				u := w[j]
				v := w[j+step]
				w[j] = u + v
				w[j+step] = u - v
			}
		}
	}
	return w
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// nonlinearityFast - то же самое, но через Уолша-Адамара.
// Формула: NL = 2^(n-1) - max|w[a]| / 2
func nonlinearityFast(f []int) int {
	w := walsh(f)
	maxW := 0
	for _, v := range w {
		if abs(v) > maxW {
			maxW = abs(v)
		}
	}
	return len(f)/2 - maxW/2
}
