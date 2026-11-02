package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

// readLine печатает приглашение и читает одну строку с клавиатуры
func readLine(prompt string) string {
	fmt.Print(prompt)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		// ввод закончился (например, Ctrl+D) - выходим
		fmt.Println()
		os.Exit(0)
	}
	return strings.TrimSpace(line)
}

// readInt читает целое число, пока не введут нормально
func readInt(prompt string) int {
	for {
		s := readLine(prompt)
		x, err := strconv.Atoi(s)
		if err == nil {
			return x
		}
		fmt.Println("  Нужно целое число, попробуйте ещё раз.")
	}
}

// parseBits достаёт из строки нули и единицы, всё остальное игнорирует.
// Поэтому можно вводить и "0110", и "0,1,1,0", и "(0 1 1 0)".
func parseBits(s string) []int {
	var bits []int
	for _, ch := range s {
		if ch == '0' {
			bits = append(bits, 0)
		} else if ch == '1' {
			bits = append(bits, 1)
		}
	}
	return bits
}

func isPowerOfTwo(x int) bool {
	return x >= 2 && x&(x-1) == 0
}

// readFunc просит ввести таблицу истинности булевой функции.
// Длина должна быть степенью двойки (2, 4, 8, 16...).
func readFunc(prompt string) []int {
	for {
		f := parseBits(readLine(prompt))
		if isPowerOfTwo(len(f)) {
			return f
		}
		fmt.Printf("  Количество значений: %d, а нужно 2^n (2, 4, 8, 16...). Ещё раз.\n", len(f))
	}
}

// bitsToString превращает [0 1 1 0] в "0110"
func bitsToString(f []int) string {
	s := ""
	for _, v := range f {
		s += strconv.Itoa(v)
	}
	return s
}

// ---------- пункты меню ----------

// пункт 1: расстояние Хэмминга
func menuHamming() {
	fmt.Println("\n--- Расстояние Хэмминга между функциями ---")
	fmt.Println("Вводите таблицы истинности, например 0001 для функции \"И\".")
	f := readFunc("Функция f: ")
	g := readFunc("Функция g: ")
	if len(f) != len(g) {
		fmt.Println("Функции должны быть от одного числа переменных (одинаковая длина).")
		return
	}
	d := distance(f, g)
	fmt.Println("f + g (XOR)  :", bitsToString(xorFunc(f, g)))
	fmt.Println("ρ(f, g)      :", d)
	fmt.Println("совпадений   :", len(f)-d)
}

// пункт 2: нелинейность функции
func menuNonlinearity() {
	fmt.Println("\n--- Нелинейность булевой функции ---")
	fmt.Println("Введите таблицу истинности (порядок аргументов: 000, 001, 010, ...).")
	f := readFunc("Функция f: ")
	n := numVars(f)
	fmt.Printf("Число переменных n = %d\n", n)

	// перебор аффинных функций долго работает для больших n
	if len(f) <= 1024 {
		fmt.Println("NL (перебор аффинных функций):", nonlinearity(f))
	} else {
		fmt.Println("NL (перебор): пропущен, функция слишком большая")
	}

	fmt.Println("NL(f + 1)                    :", nonlinearity(addOne(f)), "(должна быть такой же)")

	if n%2 == 0 {
		max := bentMaxNL(n)
		fmt.Println("Максимум для чётного n       :", max)
		if nonlinearity(f) == max {
			fmt.Println("Функция bent!")
		}
	}
}

// пункт 3: построение bent-функции
func menuBent() {
	fmt.Println("\n--- Построение bent-функции ---")
	n := readInt("Число переменных n (чётное, от 2 до 6): ")
	if n < 2 || n > 16 || n%2 != 0 {
		fmt.Println("Нужно чётное n от 2 до 6.")
		return
	}
	f := bentFunc(n)

	// выводим формулу
	half := n / 2
	formula := ""
	for i := 0; i < half; i++ {
		if i > 0 {
			formula += " + "
		}
		formula += fmt.Sprintf("x%d*x%d", i, i+half)
	}
	fmt.Println("f =", formula)

	if len(f) <= 64 {
		fmt.Println("Таблица истинности:", bitsToString(f))
	} else {
		fmt.Printf("Таблица истинности из %d значений не выводится.\n", len(f))
	}

	if len(f) <= 1024 {
		fmt.Println("NL (перебор)    :", nonlinearity(f))
	}
	fmt.Println("Теория 2^(n-1) - 2^(n/2-1) =", bentMaxNL(n))
}

func main() {
	for {
		fmt.Println()
		fmt.Println("===== ЛР2: s-боксы и линейная аппроксимация =====")
		fmt.Println("1. Расстояние Хэмминга между двумя функциями")
		fmt.Println("2. Нелинейность булевой функции")
		fmt.Println("3. Построить bent-функцию")
		fmt.Println("4. Нелинейность s-бокса")
		fmt.Println("5. Сгенерировать s-бокс размерности 5")
		fmt.Println("6. Таблица линейной аппроксимации, система уравнений, Err")
		fmt.Println("0. Выход")

		choice := readLine("Выберите пункт: ")
		switch choice {
		case "1":
			menuHamming()
		case "2":
			menuNonlinearity()
		case "3":
			menuBent()
		case "4":
			menuSbox()
		case "5":
			menuGenerate()
		case "6":
			menuLAT()
		case "0":
			fmt.Println("До свидания!")
			return
		default:
			fmt.Println("Нет такого пункта.")
		}
	}
}
