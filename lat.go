package main

import (
	"fmt"
	"sort"
)

// buildLAT строит таблицу линейной аппроксимации s-бокса.
//
//	T[b][a] = (сколько x удовлетворяют a*X = b*Y, где Y = S(x)) - 2^(n-1)
//
// Строки - маски выходных битов b (их 2^m), столбцы - маски входных a (их 2^n).
// Маски записаны так же, как и всё остальное: старший бит отвечает за X0 (или Y0).
// Скалярное произведение по модулю 2 считаем как parity(a & x).
func buildLAT(s []int, n, m int) [][]int {
	T := make([][]int, 1<<m)
	for b := range T {
		T[b] = make([]int, 1<<n)
		for a := range T[b] {
			count := 0
			for x := range s {
				if parity(a&x) == parity(b&s[x]) {
					count++
				}
			}
			T[b][a] = count - (1 << (n - 1))
		}
	}
	return T
}

// latMax ищет наибольшее по модулю значение в таблице (кроме строки b = 0,
// она всегда даёт 2^(n-1) в первой клетке и нули дальше, это не информация).
// Возвращает само значение |T| и список клеток [b, a], где оно достигается.
func latMax(T [][]int) (int, [][2]int) {
	best := 0
	var cells [][2]int
	for b := 1; b < len(T); b++ {
		for a := range T[b] {
			v := abs(T[b][a])
			if v > best {
				best = v
				cells = nil
			}
			if v == best {
				cells = append(cells, [2]int{b, a})
			}
		}
	}
	return best, cells
}

// latNonlinearity - нелинейность s-бокса по таблице: 2^(n-1) - max|T|
func latNonlinearity(T [][]int, n int) int {
	best, _ := latMax(T)
	return (1 << (n - 1)) - best
}

// approx - лучшая линейная аппроксимация одной базисной функции Yj:
// Yj = a*X + c
type approx struct {
	j      int   // номер выхода
	a      int   // маска коэффициентов при X0..X(n-1)
	c      int   // свободный член (0 или 1)
	t      int   // значение в таблице LAT
	others []int // другие столбцы с таким же |T| (на выбор, если захочется)
	errs   []int // аргументы x, на которых аппроксимация ошибается
}

// bestApprox выбирает для выхода Yj наибольший по модулю элемент в его строке.
// Строка для Yj - это маска b с единицей только в позиции j.
// Если элемент отрицательный, значит равенство a*X = Yj чаще не выполняется,
// и в уравнение добавляется +1.
func bestApprox(T [][]int, s []int, n, m, j int) approx {
	b := 1 << (m - 1 - j)
	res := approx{j: j}

	// ищем максимум |T| в строке
	maxAbs := -1
	for a := range T[b] {
		if abs(T[b][a]) > maxAbs {
			maxAbs = abs(T[b][a])
			res.a = a
		}
	}
	res.t = T[b][res.a]
	if res.t < 0 {
		res.c = 1
	}

	// запоминаем остальные столбцы с таким же модулем
	for a := range T[b] {
		if a != res.a && abs(T[b][a]) == maxAbs {
			res.others = append(res.others, a)
		}
	}

	// находим аргументы, где уравнение не выполняется
	for x := range s {
		y := (s[x] >> (m - 1 - j)) & 1
		if y != parity(res.a&x)^res.c {
			res.errs = append(res.errs, x)
		}
	}
	return res
}

// eqString записывает уравнение в виде строки, например "Y3 = X0 + X3"
func eqString(ap approx, n int) string {
	terms := ""
	for i := 0; i < n; i++ {
		if (ap.a>>(n-1-i))&1 == 1 {
			if terms != "" {
				terms += " + "
			}
			terms += fmt.Sprintf("X%d", i)
		}
	}
	if ap.c == 1 {
		if terms == "" {
			terms = "1"
		} else {
			terms += " + 1"
		}
	}
	if terms == "" {
		terms = "0"
	}
	return fmt.Sprintf("Y%d = %s", ap.j, terms)
}

// allApprox строит лучшие аппроксимации сразу для всех выходов Y0..Y(m-1)
func allApprox(T [][]int, s []int, n, m int) []approx {
	var res []approx
	for j := 0; j < m; j++ {
		res = append(res, bestApprox(T, s, n, m, j))
	}
	return res
}

// errSet - объединение множеств ошибок всех уравнений (по возрастанию)
func errSet(aps []approx) []int {
	seen := map[int]bool{}
	var res []int
	for _, ap := range aps {
		for _, x := range ap.errs {
			if !seen[x] {
				seen[x] = true
				res = append(res, x)
			}
		}
	}
	sort.Ints(res)
	return res
}
