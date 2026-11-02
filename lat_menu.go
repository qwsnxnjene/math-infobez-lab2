package main

import "fmt"

// printLAT выводит таблицу линейной аппроксимации.
// Столбцы - маски входов a, строки - маски выходов b.
func printLAT(T [][]int, n, m int) {
	fmt.Println("\nТаблица линейной аппроксимации T[b][a]:")
	fmt.Println("  столбец a - маска входа (X0 старший бит), строка b - маска выхода (Y0 старший бит)")
	fmt.Println("  например, a = 9 = 1001 это X0 + X3, а b = 3 = 0011 это Y2 + Y3")
	fmt.Println()

	fmt.Printf("  %-17s|", "a ->")
	for a := 0; a < (1 << n); a++ {
		fmt.Printf("%4d", a)
	}
	fmt.Println()
	fmt.Print("  -----------------+")
	for a := 0; a < (1 << n); a++ {
		fmt.Print("----")
	}
	fmt.Println()

	for b := range T {
		label := combName(b, m)
		if b == 0 {
			label = "0"
		}
		fmt.Printf("  %2d %-14s|", b, label)
		for a := range T[b] {
			fmt.Printf("%4d", T[b][a])
		}
		fmt.Println()
	}
}

// checkRelation показывает, как проверяется одно соотношение a*X = b*Y
// на всех аргументах (как таблицы 8 и 10 в пособии)
func checkRelation(s []int, T [][]int, n, m, a, b int) {
	fmt.Printf("\nСоотношение: %s = %s\n", maskName(a, n, "X"), maskName(b, m, "Y"))
	if n <= 4 {
		fmt.Println("  x    S(x)   a*X  b*Y  совпало?")
		for x := range s {
			left := parity(a & x)
			right := parity(b & s[x])
			mark := "да"
			if left != right {
				mark = "нет"
			}
			fmt.Printf("  %-4d %-6d %-4d %-4d %s\n", x, s[x], left, right, mark)
		}
	}
	matches := T[b][a] + (1 << (n - 1))
	fmt.Printf("  Совпадений: %d из %d, T = %d - %d = %d\n", matches, len(s), matches, 1<<(n-1), T[b][a])
}

// maskName записывает маску как сумму переменных, например "X0 + X3"
func maskName(mask, bits int, letter string) string {
	name := ""
	for i := 0; i < bits; i++ {
		if (mask>>(bits-1-i))&1 == 1 {
			if name != "" {
				name += " + "
			}
			name += fmt.Sprintf("%s%d", letter, i)
		}
	}
	if name == "" {
		name = "0"
	}
	return name
}

// пункт 7: LAT, система уравнений и множество ошибок
func menuLAT() {
	fmt.Println("\n--- Таблица линейной аппроксимации s-бокса ---")
	s, n, m, ok := readSbox()
	if !ok {
		return
	}
	if n > 8 {
		fmt.Println("Слишком большой s-бокс.")
		return
	}

	fmt.Printf("\ns-бокс (%d, %d)\n", n, m)
	fmt.Print("S(x) = ")
	fmt.Println(s)

	T := buildLAT(s, n, m)
	if n <= 5 && m <= 5 {
		printLAT(T, n, m)
	} else {
		fmt.Println("\nТаблица слишком большая, на экран не выводится.")
	}

	// нелинейность двумя способами
	maxT, cells := latMax(T)
	nlLAT := latNonlinearity(T, n)
	nlWalsh := sboxNonlinearity(s, m)
	fmt.Println("\nНаибольшее |T| (строка b = 0 не учитывается):", maxT)
	fmt.Print("Достигается в клетках (b, a): ")
	for i, c := range cells {
		if i >= 8 {
			fmt.Printf("... (всего %d)", len(cells))
			break
		}
		fmt.Printf("(%d, %d) ", c[0], c[1])
	}
	fmt.Println()
	fmt.Printf("Нелинейность по LAT: 2^(n-1) - max|T| = %d - %d = %d\n", 1<<(n-1), maxT, nlLAT)
	fmt.Printf("Нелинейность через Уолша-Адамара: %d\n", nlWalsh)
	if nlLAT != nlWalsh {
		fmt.Println("ВНИМАНИЕ: значения не совпали!")
	}

	// система уравнений: лучшая аппроксимация для каждого Yj
	aps := allApprox(T, s, n, m)
	fmt.Println("\nСистема линейных уравнений (лучшая аппроксимация каждого Yj):")
	for _, ap := range aps {
		matches := (1 << (n - 1)) + abs(ap.t)
		fmt.Printf("  %-22s T = %3d, верно на %d из %d, ошибка при x = %v\n",
			eqString(ap, n), ap.t, matches, len(s), ap.errs)
		if len(ap.others) > 0 {
			fmt.Printf("      (столбцы с таким же |T|: %v, взят первый)\n", ap.others)
		}
	}

	errs := errSet(aps)
	fmt.Println("\nМножество Err (где хотя бы одно уравнение ошибается):", errs)
	if n <= 5 {
		var good []int
		for x := range s {
			isErr := false
			for _, e := range errs {
				if e == x {
					isErr = true
				}
			}
			if !isErr {
				good = append(good, x)
			}
		}
		fmt.Println("Аргументы без ошибок:", good)
	}

	// по желанию проверяем ещё какое-нибудь соотношение вручную
	fmt.Println()
	ab := parseInts(readLine("Проверить своё соотношение a*X = b*Y? Введите a и b (Enter - пропустить): "))
	if len(ab) == 2 && ab[0] < (1<<n) && ab[1] < (1<<m) {
		checkRelation(s, T, n, m, ab[0], ab[1])
	} else if len(ab) > 0 {
		fmt.Printf("Нужны два числа: a от 0 до %d и b от 0 до %d.\n", (1<<n)-1, (1<<m)-1)
	}
}
