package utils

import (
	"bufio"
	"fmt"
	"log"
	"strconv"

	"github.com/fatih/color"
)

// RequestParameter Запросить у пользователя параметр в виде строки
func RequestParameter(msg string, required bool, scanner *bufio.Scanner) string {
	for {
		fmt.Printf(color.BlueString(msg) + ": ")
		if !scanner.Scan() {
			if scanner.Err() == nil {
				log.Fatalf("Ввод из консоли принудительно завершен")
			} else {
				fmt.Println(color.YellowString("Не удалось прочитать из консоли"))
				continue
			}
		}
		parameter := scanner.Text()
		if required && parameter == "" {
			fmt.Println(color.YellowString("Этот параметр является обязательным"))
			continue
		}
		return parameter
	}
}

// RequestChoice Запросить у пользователя выбор строки из предложенных строк
func RequestChoice(msg string, a []string, scanner *bufio.Scanner) int {
	if len(a) <= 0 {
		log.Fatalf("Ошибка, передано 0 возможных значений")
	}
	for i, aa := range a {
		fmt.Printf("%d. %s\n", i, aa)
	}
	for {
		input := RequestParameter(msg, true, scanner)
		n, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println(color.YellowString("Ошибка конвертации в целое число"))
			continue
		}
		if n < 0 || n >= len(a) {
			fmt.Println(color.YellowString("Введите число в промежутке [%d, %d]", 0, len(a)-1))
			continue
		}
		return n
	}
}
