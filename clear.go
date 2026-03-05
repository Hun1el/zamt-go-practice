package main

import (
	"fmt"
	"os"
	"strings"
)

// ClearFlags содержит флаги для управления clear
type ClearFlags struct {
	Terminal string // -T указать тип терминала (TERM)
	Version  bool   // -V показать версию
	Noclear  bool   // -x не очищать экран, а использовать метод со смещением
	Help     bool   // -h/--help справка
}

// ShowHelp выводит справку по использованию команды clear
func ShowHelp() {
	help := `
Использование: ./clear [КЛЮЧИ]

Ключи:
  -T TERM    Установить тип терминала
  -V         Показать версию программы
  -x         Использовать метод со смещением курсора
  -h         Показать эту справку
  --help     Показать эту справку

Описание:
  clear - очищает экран терминала. 
  Команда clear используется для очистки содержимого экрана терминала.
  Удаляет всё содержимое, которое было отображено ранее, и помещает
  курсор в верхний левый угол экрана.

Примеры:
  ./clear                Очистить экран
  ./clear -V             Показать версию
  ./clear -T xterm       Очистить экран xterm
  ./clear -x             Прокрутка вверх
  ./clear -T linux -x    Linux терминал с методом прокрутки
`
	fmt.Println(help)
}

// ParseClearFlags парсит флаги для clear
func ParseClearFlags(args []string) (ClearFlags, error) {
	flags := ClearFlags{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			flags.Help = true
		} else if arg == "-V" {
			flags.Version = true
		} else if arg == "-x" {
			flags.Noclear = true
		} else if arg == "-T" {
			if i+1 < len(args) {
				flags.Terminal = args[i+1]
				i++
			} else {
				return flags, fmt.Errorf("ошибка: флаг -T требует значение (тип терминала)")
			}
		} else if strings.HasPrefix(arg, "-") && arg != "-" {
			return flags, fmt.Errorf("неизвестный флаг: %s", arg)
		}
	}

	return flags, nil
}

// ClearScreen очищает экран терминала
func ClearScreen(flags ClearFlags) {
	// Метод прокрутки
	if flags.Noclear {
		for i := 0; i < 36; i++ {
			fmt.Println()
		}
		// переместить курсор в начало
		fmt.Fprint(os.Stdout, "\033[H")
		os.Stdout.Sync()
		return
	}

	// Получаем тип терминала
	termType := flags.Terminal
	if termType == "" {
		termType = os.Getenv("TERM")
	}

	if termType == "" || termType == "dumb" {
		for i := 0; i < 36; i++ {
			fmt.Println()
		}
		fmt.Fprint(os.Stdout, "\033[H")
		os.Stdout.Sync()
		return
	}

	fmt.Fprint(os.Stdout, "\033[H")  // Переместить курсор в начало
	fmt.Fprint(os.Stdout, "\033[J")  // Очистить от текущей позиции до конца
	fmt.Fprint(os.Stdout, "\033[3J") // Очистить весь буфер прокрутки
	os.Stdout.Sync()
}

func main() {
	flags, err := ParseClearFlags(os.Args[1:])
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("Используй ./clear --help для справки")
		os.Exit(1)
	}

	if flags.Help {
		ShowHelp()
		os.Exit(0)
	}

	if flags.Version {
		fmt.Println("Clear version 1.0.0 (Go implementation)")
		fmt.Println("Clears the terminal screen")
		os.Exit(0)
	}

	ClearScreen(flags)
	os.Exit(0)
}
