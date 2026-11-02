package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// S-бокс хранится как срез s длины 2^n: s[x] - значение s-бокса на аргументе x
// (число от 0 до 2^m - 1). Биты значения нумеруются так же, как и у аргумента:
// y0 - старший бит. Для перестановок n = m.

// sboxBasis возвращает базисные функции s-бокса: Y0, Y1, ... Y(m-1).
// Yj(x) - это j-й бит значения s[x].
func sboxBasis(s []int, m int) [][]int {
	basis := make([][]int, m)
	for j := 0; j < m; j++ {
		basis[j] = make([]int, len(s))
		for x := range s {
			basis[j][x] = (s[x] >> (m - 1 - j)) & 1
		}
	}
	return basis
}

// sboxFromBasis - обратная операция: собирает s-бокс из базисных функций
func sboxFromBasis(basis [][]int) []int {
	m := len(basis)
	s := make([]int, len(basis[0]))
	for x := range s {
		for j := 0; j < m; j++ {
			s[x] = s[x]*2 + basis[j][x]
		}
	}
	return s
}

// sboxComb - линейная комбинация базисных функций b0*Y0 + b1*Y1 + ...
// Маска b записана битами так же, как значение (b0 - старший бит).
// Это функция от x: parity(b & s[x])
func sboxComb(s []int, b int) []int {
	f := make([]int, len(s))
	for x := range s {
		f[x] = parity(b & s[x])
	}
	return f
}

// sboxNonlinearity - нелинейность s-бокса: наименьшая нелинейность среди
// всех ненулевых линейных комбинаций базисных функций.
// (комбинация с b = 0 - это константа 0, её не берём)
func sboxNonlinearity(s []int, m int) int {
	best := len(s)
	for b := 1; b < (1 << m); b++ {
		nl := nonlinearity(sboxComb(s, b))
		if nl < best {
			best = nl
		}
	}
	return best
}

// isBijection проверяет, что все значения s-бокса разные (перестановка)
func isBijection(s []int) bool {
	seen := make([]bool, len(s))
	for _, v := range s {
		if v < 0 || v >= len(s) || seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}

// fixedPoints возвращает аргументы, на которых s(x) = x
func fixedPoints(s []int) []int {
	var res []int
	for x := range s {
		if s[x] == x {
			res = append(res, x)
		}
	}
	return res
}

// ---------- s-бокс через инверсию в поле GF(2^n) ----------

// неприводимые многочлены для полей GF(2^n), записанные битами
// (например, для n=5: x^5 + x^2 + 1 = 100101 = 37)
var gfPoly = map[int]int{
	3: 0xB,   // x^3 + x + 1
	4: 0x13,  // x^4 + x + 1
	5: 0x25,  // x^5 + x^2 + 1
	6: 0x43,  // x^6 + x + 1
	7: 0x83,  // x^7 + x + 1
	8: 0x11B, // x^8 + x^4 + x^3 + x + 1 (как в AES)
}

// gfMul умножает два элемента поля GF(2^n)
func gfMul(a, b, n int) int {
	res := 0
	for b > 0 {
		if b&1 == 1 {
			res ^= a
		}
		b >>= 1
		a <<= 1
		if a&(1<<n) != 0 { // степень вышла за n - сокращаем по многочлену
			a ^= gfPoly[n]
		}
	}
	return res
}

// powerSbox строит s-бокс x -> x^k в поле GF(2^n) (0 переходит в 0).
func powerSbox(n, k int) []int {
	s := make([]int, 1<<n)
	for x := 1; x < len(s); x++ {
		r := 1
		for i := 0; i < k; i++ {
			r = gfMul(r, x, n)
		}
		s[x] = r
	}
	return s
}

// removeFixedPoints прибавляет к выходу s-бокса константу c так, чтобы
// не осталось неподвижных точек.
// Возвращает новый s-бокс и найденную константу; ok=false, если не вышло.
func removeFixedPoints(s []int) ([]int, int, bool) {
	for c := 0; c < len(s); c++ {
		res := make([]int, len(s))
		for x := range s {
			res[x] = s[x] ^ c
		}
		if len(fixedPoints(res)) == 0 {
			return res, c, true
		}
	}
	return s, 0, false
}

// ---------- случайный поиск s-бокса (hill climbing) ----------

// штраф за каждую неподвижную точку (очень большой, чтобы от них избавиться)
const fixedPenalty = 1 << 42

// sboxCost - "стоимость" s-бокса, чем меньше, тем лучше.
// Берём сумму |W|^8 по спектрам Уолша всех линейных комбинаций:
// большие пики в спектре (то есть близость к линейной функции) очень дорого стоят.
func sboxCost(s []int, n int) int {
	total := 0
	for b := 1; b < (1 << n); b++ {
		w := walsh(sboxComb(s, b))
		for _, v := range w {
			a := abs(v)
			p := 1
			for i := 0; i < 8; i++ {
				p *= a
			}
			total += p
		}
	}
	return total + fixedPenalty*len(fixedPoints(s))
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

// randomSearch ищет s-бокс (перестановку без неподвижных точек) с большой
// нелинейностью. Начинаем со случайной перестановки и много раз меняем
// местами два значения: если стоимость не выросла, оставляем обмен,
// иначе возвращаем всё как было. Возвращает лучший найденный s-бокс.
func randomSearch(n, iters int, rnd *rand.Rand) []int {
	s := rnd.Perm(1 << n)
	cur := sboxCost(s, n)

	var best []int
	bestNL := -1
	if len(fixedPoints(s)) == 0 {
		best = append([]int{}, s...)
		bestNL = sboxNonlinearity(s, n)
		fmt.Printf("  старт: NL = %d\n", bestNL)
	}

	for it := 1; it <= iters; it++ {
		i := rnd.Intn(len(s))
		j := rnd.Intn(len(s))
		if i == j {
			continue
		}
		s[i], s[j] = s[j], s[i]
		c := sboxCost(s, n)
		if c > cur {
			s[i], s[j] = s[j], s[i] // стало хуже - откатываем
			continue
		}
		if c < cur && len(fixedPoints(s)) == 0 {
			nl := sboxNonlinearity(s, n)
			if nl > bestNL {
				bestNL = nl
				best = append([]int{}, s...)
				fmt.Printf("  итерация %d: NL = %d\n", it, nl)
			}
		}
		cur = c
	}
	return best
}

// parseInts достаёт из строки числа, всё остальное считается разделителем.
// Подойдёт и "1 0 3 2", и "1,0,3,2".
func parseInts(s string) []int {
	var res []int
	parts := strings.FieldsFunc(s, func(r rune) bool { return r < '0' || r > '9' })
	for _, p := range parts {
		x, _ := strconv.Atoi(p)
		res = append(res, x)
	}
	return res
}

// combName - название линейной комбинации по маске b, например "Y0+Y2"
func combName(b, m int) string {
	name := ""
	for j := 0; j < m; j++ {
		if (b>>(m-1-j))&1 == 1 {
			if name != "" {
				name += "+"
			}
			name += fmt.Sprintf("Y%d", j)
		}
	}
	return name
}

// printSboxInfo выводит всё про s-бокс: таблицу, базисные функции,
// нелинейность каждой комбинации и итоговую нелинейность.
// n - число входных бит, m - число выходных.
func printSboxInfo(s []int, n, m int) {
	fmt.Println("\nТаблица подстановки:")
	fmt.Print("  x    :")
	for x := range s {
		fmt.Printf("%3d", x)
	}
	fmt.Print("\n  S(x) :")
	for _, v := range s {
		fmt.Printf("%3d", v)
	}
	fmt.Println()

	fmt.Println("\nБазисные функции:")
	basis := sboxBasis(s, m)
	for j := 0; j < m; j++ {
		fmt.Printf("  Y%d (F%d) = %s\n", j, j+1, bitsToString(basis[j]))
	}

	fmt.Println("\nНелинейность линейных комбинаций:")
	for b := 1; b < (1 << m); b++ {
		nl := nonlinearity(sboxComb(s, b))
		fmt.Printf("  %-14s NL = %d\n", combName(b, m), nl)
	}

	fmt.Println()
	fmt.Println("Нелинейность s-бокса (минимум по комбинациям):", sboxNonlinearity(s, m))

	// биекция и неподвижные точки имеют смысл только при n = m
	if n == m {
		if isBijection(s) {
			fmt.Println("Биекция (перестановка): да")
		} else {
			fmt.Println("Биекция (перестановка): нет")
		}
		fp := fixedPoints(s)
		if len(fp) == 0 {
			fmt.Println("Неподвижные точки     : нет")
		} else {
			fmt.Println("Неподвижные точки     :", fp)
		}
	}
}

// readSbox спрашивает, как задан s-бокс, и считывает его.
// Возвращает таблицу подстановки s, число входных бит n и выходных m.
// ok = false, если ввод оказался неправильным.
func readSbox() (s []int, n, m int, ok bool) {
	fmt.Println("Как задан s-бокс?")
	fmt.Println("  1. Таблицей подстановки: S(0), S(1), S(2), ...")
	fmt.Println("  2. Базисными функциями Y0, Y1, ... (таблицы истинности)")
	fmt.Println("  3. Вариант 14 (F1..F4 из задания)")
	fmt.Println("  4. Пример 3x3 из пособия (ЛР2, таблицы 8-10)")
	fmt.Println("  5. Пример 4x4 из пособия (ЛР3)")
	mode := readLine("Выбор: ")

	switch mode {
	case "1":
		s = parseInts(readLine("S: "))
		if !isPowerOfTwo(len(s)) {
			fmt.Println("Количество значений должно быть степенью двойки.")
			return nil, 0, 0, false
		}
		n = numVars(s)
		m = n // для таблицы подстановки считаем, что вход и выход одной длины
		for _, v := range s {
			if v >= len(s) {
				fmt.Printf("Значения должны быть от 0 до %d.\n", len(s)-1)
				return nil, 0, 0, false
			}
		}
	case "2":
		m = readInt("Сколько базисных функций (m)? ")
		if m < 1 || m > 8 {
			fmt.Println("m должно быть от 1 до 8.")
			return nil, 0, 0, false
		}
		var basis [][]int
		for j := 0; j < m; j++ {
			basis = append(basis, readFunc(fmt.Sprintf("Y%d: ", j)))
			if len(basis[j]) != len(basis[0]) {
				fmt.Println("Все функции должны быть одинаковой длины.")
				return nil, 0, 0, false
			}
		}
		s = sboxFromBasis(basis)
		n = numVars(s)
	case "3":
		s = sboxFromBasis(variant14)
		n = numVars(s)
		m = len(variant14)
	case "4":
		s = manualSbox3
		n, m = 3, 3
	case "5":
		s = sboxFromBasis(manualBasis4)
		n, m = 4, 4
	default:
		fmt.Println("Нет такого варианта.")
		return nil, 0, 0, false
	}
	return s, n, m, true
}

// пункт 5: нелинейность заданного s-бокса
func menuSbox() {
	fmt.Println("\n--- Нелинейность s-бокса ---")
	s, n, m, ok := readSbox()
	if !ok {
		return
	}
	fmt.Printf("\ns-бокс (%d, %d)\n", n, m)
	printSboxInfo(s, n, m)
}

// пункт 6: генерация s-бокса размерности 5
func menuGenerate() {
	const n = 5
	fmt.Println("\n--- Генерация s-бокса размерности n = 5 ---")
	fmt.Println("Способ:")
	fmt.Println("  1. Случайный поиск (hill climbing)")
	fmt.Println("  2. Степенная функция x^k в поле GF(2^5)")
	mode := readLine("Выбор: ")

	switch mode {
	case "1":
		iters := readInt("Число итераций (например, 100000, это несколько секунд): ")
		if iters < 1 {
			fmt.Println("Нужно положительное число.")
			return
		}
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		fmt.Println("Идёт поиск...")
		s := randomSearch(n, iters, rnd)
		if s == nil {
			fmt.Println("Не нашлось подходящей перестановки, попробуйте ещё раз.")
			return
		}

		printSboxInfo(s, n, n)

	case "2":
		k := readInt("Степень k (3 или 5 работают лучше): ")
		if k < 1 {
			fmt.Println("Нужно k >= 1.")
			return
		}
		s := powerSbox(n, k)
		if !isBijection(s) {
			fmt.Printf("x^%d в GF(2^5) не биекция (k должно быть взаимно просто с 31).\n", k)
			return
		}
		fmt.Printf("\ns(x) = x^%d в GF(2^5), неприводимый многочлен x^5 + x^2 + 1\n", k)
		fmt.Println("Неподвижные точки до исправления:", fixedPoints(s))

		s2, c, ok := removeFixedPoints(s)
		if !ok {
			fmt.Println("Не удалось подобрать константу, неподвижные точки остались.")
		} else {
			fmt.Printf("Прибавляем к выходу константу c = %d (XOR), нелинейность от этого не меняется.\n", c)
		}
		printSboxInfo(s2, n, n)

	default:
		fmt.Println("Нет такого варианта.")
	}
}
