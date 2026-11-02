package main

import (
	"fmt"
	"math/rand"
)

// check печатает результат одной проверки
func check(name string, ok bool) {
	if ok {
		fmt.Println("  [OK]     ", name)
	} else {
		fmt.Println("  [ОШИБКА] ", name)
	}
}

// randomFunc - случайная булева функция от n переменных
func randomFunc(rnd *rand.Rand, n int) []int {
	f := make([]int, 1<<n)
	for i := range f {
		f[i] = rnd.Intn(2)
	}
	return f
}

// runChecks - контрольные примеры из пособия (пункт 4 меню)
func runChecks() {
	rnd := rand.New(rand.NewSource(1)) // фиксированное зерно, чтобы результаты не менялись

	fmt.Println("\n--- Контрольные примеры из пособия ---")

	// ЛР1: расстояние и нелинейность
	fmt.Println("Булевы функции:")
	and := []int{0, 0, 0, 1}
	or := []int{0, 1, 1, 1}
	check("H(101) = 2", hammingWeight(5) == 2)
	check("H(1010 + 1100) = 2", hammingWeight(10^12) == 2)
	check("ρ(И, ИЛИ) = 2", distance(and, or) == 2)
	check("NL(x1*x2) = 1", nonlinearity(and) == 1)
	check("NL(x0+x1+1) = 0 (аффинная функция)", nonlinearity(affine(2, 3, 1)) == 0)

	// NL(f) = NL(f+1) на случайных функциях
	ok := true
	for i := 0; i < 50; i++ {
		f := randomFunc(rnd, 4)
		if nonlinearity(f) != nonlinearity(addOne(f)) {
			ok = false
		}
	}
	check("NL(f) = NL(f+1) на 50 случайных функциях", ok)

	// перебор и Уолш-Адамар должны совпадать
	ok = true
	for n := 2; n <= 8; n++ {
		for i := 0; i < 20; i++ {
			f := randomFunc(rnd, n)
			if nonlinearity(f) != nonlinearityFast(f) {
				ok = false
			}
		}
	}
	check("перебор = Уолш-Адамар (n от 2 до 8)", ok)

	// bent-функции, таблица 4
	expected := map[int]int{2: 1, 4: 6, 6: 28, 8: 120, 10: 496}
	for n := 2; n <= 10; n += 2 {
		nl := nonlinearityFast(bentFunc(n))
		check(fmt.Sprintf("bent n=%d: NL = %d (в таблице 4: %d)", n, nl, expected[n]), nl == expected[n])
	}

	// s-боксы
	fmt.Println("S-боксы:")

	// сдвиг (x1,x2,x3) -> (x2,x3,x1) из таблицы 5: все функции - проекции
	shift := []int{0, 2, 4, 6, 1, 3, 5, 7}
	check("s-бокс-сдвиг (таблица 5) имеет NL = 0", sboxNonlinearity(shift, 3) == 0)

	// s-бокс 3x3 из ЛР2 (версия из таблиц 8 и 10, по ней построена таблица 9)
	nl3 := sboxNonlinearity(manualSbox3, 3)
	check(fmt.Sprintf("s-бокс 3x3 из ЛР2: NL = %d (в пособии 1)", nl3), nl3 == 1)

	// s-бокс 4x4 из ЛР3, в его таблице LAT максимум |T| = 7, значит NL = 8 - 7 = 1
	lab3 := sboxFromBasis(manualBasis4)
	nl4 := sboxNonlinearity(lab3, 4)
	check(fmt.Sprintf("s-бокс из ЛР3: NL = %d (в LAT максимум 7, значит 8 - 7 = 1)", nl4), nl4 == 1)

	// известные значения для степенных s-боксов
	check("инверсия в GF(2^8) (как в AES): NL = 112", sboxNonlinearity(powerSbox(8, 254), 8) == 112)
	check("x^3 в GF(2^5): биекция, NL = 12", isBijection(powerSbox(5, 3)) && sboxNonlinearity(powerSbox(5, 3), 5) == 12)

	// итоговый s-бокс для n=5
	s5, _, found := removeFixedPoints(powerSbox(5, 3))
	check("s-бокс n=5: биекция, без неподвижных точек, NL = 12",
		found && isBijection(s5) && len(fixedPoints(s5)) == 0 && sboxNonlinearity(s5, 5) == 12)

	// таблица линейной аппроксимации
	fmt.Println("LAT:")

	// таблица 9 из ЛР2. В пособии в строке (1,0,1) опечатка в двух клетках
	// (напечатано 1 и 0), мы ставим правильные -1 и 1: так сумма квадратов
	// в строке равна 16, как должно быть по теореме Парсеваля.
	table9 := [][]int{
		{4, 0, 0, 0, 0, 0, 0, 0},
		{-1, 3, -1, -1, 1, 1, 1, 1},
		{0, -2, 0, 2, 2, 0, 2, 0},
		{-1, 1, 3, 1, -1, 1, -1, 1},
		{-2, -2, 0, 0, 0, 0, -2, 2},
		{-1, -1, 1, 1, -1, -1, 1, -3},
		{0, 2, -2, 0, 0, -2, -2, 0},
		{1, -1, -1, -3, -1, 1, 1, -1},
	}
	check("LAT 3x3 совпадает с таблицей 9 (с исправлением 2 опечаток)",
		sameTable(buildLAT(manualSbox3, 3, 3), table9))

	// таблица LAT из ЛР3 (4x4), в пособии напечатана без ошибок
	table16 := [][]int{
		{8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{1, 1, -1, -1, 1, 1, -1, -1, -1, 7, 1, 1, -1, -1, 1, 1},
		{1, -1, -1, 1, -1, 1, 1, 7, -1, 1, 1, -1, 1, -1, -1, 1},
		{0, -2, 0, 2, -2, 0, 2, 0, 0, 2, 0, -2, 2, 0, 6, 0},
		{1, 7, 1, -1, 1, -1, 1, -1, 1, -1, 1, -1, 1, -1, 1, -1},
		{0, 2, -2, 0, 0, 2, -2, 0, 6, 0, 0, 2, -2, 0, 0, 2},
		{-2, 2, 0, 0, 0, 0, 6, 2, 0, 0, -2, 2, -2, 2, 0, 0},
		{-1, -1, 3, -1, 1, -3, 1, 1, 3, -1, -1, -1, 1, 1, 1, 5},
		{-1, 1, -1, 1, -7, -1, 1, -1, -1, 1, -1, 1, 1, -1, 1, -1},
		{0, -2, 2, 0, -2, 0, 0, 2, 2, 0, 0, -2, 0, -6, -2, 0},
		{0, 0, -2, -6, 0, 0, 2, -2, -2, 2, 0, 0, 2, -2, 0, 0},
		{3, -1, -1, -1, -1, 3, -1, -1, -1, -1, -5, -1, -1, -1, -1, 3},
		{0, 0, 0, 0, -2, -6, -2, 2, 0, 0, 0, 0, -2, 2, -2, 2},
		{-1, -1, 1, 1, 1, -3, 3, -1, 1, 1, -1, -1, -5, -1, 1, -3},
		{1, -1, -5, -3, 1, -1, -1, 1, 3, -3, 1, -1, -1, 1, 1, -1},
		{-2, 4, -2, 0, 2, 0, -2, 0, -2, 0, -2, -4, -2, 0, 2, 0},
	}
	T4 := buildLAT(lab3, 4, 4)
	check("LAT 4x4 совпадает с таблицей из ЛР3 (все 256 клеток)", sameTable(T4, table16))
	check("нелинейность по LAT = нелинейность по Уолшу (ЛР3)", latNonlinearity(T4, 4) == sboxNonlinearity(lab3, 4))

	// система уравнений (3.1) и набор Err = {5, 1, 15, 10} из ЛР3
	aps := allApprox(T4, lab3, 4, 4)
	wantEq := []string{"Y0 = X1 + 1", "Y1 = X3", "Y2 = X1 + X2 + X3", "Y3 = X0 + X3"}
	ok = true
	for j, ap := range aps {
		if eqString(ap, 4) != wantEq[j] {
			ok = false
		}
	}
	check("система уравнений совпадает с (3.1) из ЛР3", ok)
	wantErr := []int{5, 1, 15, 10}
	ok = true
	for j, ap := range aps {
		if len(ap.errs) != 1 || ap.errs[0] != wantErr[j] {
			ok = false
		}
	}
	check("ошибки уравнений: x = 5, 1, 15, 10, как в пособии", ok)
}

// sameTable сравнивает две таблицы целиком
func sameTable(a, b [][]int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}
