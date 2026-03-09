package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CatFlags содержит флаги для управления выводом cat
type CatFlags struct {
	LineNumbers bool // -n показывать номера строк
	ShowAll     bool // -A показывать все непечатные символы
	NumberBlank bool // -b нумеровать только непустые строки
	Help        bool // -h/--help справка
}

// ShowHelp выводит справку по использованию команды cat
func ShowHelp() {
	help := `
Использование: ./cat [КЛЮЧИ] [FILE1] [FILE2] [FILE3] ...

Ключи:
  -n         Нумеровать все строки (включая пустые)
  -b         Нумеровать только непустые строки
  -A         Показывать все непечатные символы ($ в конце строк, ^I для табов)
  -h         Показать эту справку
  --help     Показать эту справку

Описание:
  cat - выводит содержимое файлов на стандартный вывод.
  Используется для просмотра содержимого файлов.
  Если файлы не указаны, читает из стандартного входа.
  Может выводить один или несколько файлов подряд, нумеровать строки
  и показывать непечатные символы.

Примеры:
  ./cat file.txt                 Вывести содержимое файла
  ./cat                          Читать со стандартного входа
  ./cat -n file.txt              Вывести с нумерацией всех строк
  ./cat -n                       Читать со stdin с нумерацией
  ./cat -b file.txt              Вывести с нумерацией только непустых строк
  ./cat -A file.txt              Показать непечатные символы
  ./cat -nb file.txt             Нумерация только непустых строк (эквивалент -b)
  ./cat -nA file.txt             Нумерация всех строк с показом символов
  ./cat file1.txt file2.txt      Вывести несколько файлов
  ./cat -n file1.txt file2.txt   Несколько файлов с нумерацией
`
	fmt.Println(help)
}

// ParseCatFlags парсит флаги для cat
func ParseCatFlags(args []string) (CatFlags, []string, error) {
	flags := CatFlags{}
	var files []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			flags.Help = true
		} else if strings.HasPrefix(arg, "-") && arg != "-" {
			// Обработка комбинированных флагов
			for j := 1; j < len(arg); j++ {
				ch := rune(arg[j])
				switch ch {
				case 'n':
					flags.LineNumbers = true
				case 'A':
					flags.ShowAll = true
				case 'b':
					flags.NumberBlank = true
				case 'h':
					flags.Help = true
				default:
					return flags, nil, fmt.Errorf("неизвестный флаг: -%c", ch)
				}
			}
		} else {
			files = append(files, arg)
		}
	}

	return flags, files, nil
}

// VisibleizeNonPrintable преобразует непечатные символы в видимые представления
func VisibleizeNonPrintable(ch rune) string {
	switch ch {
	case '\t':
		return "^I"
	case '\n':
		return "" // Новая строка обрабатывается отдельно
	case '\r':
		return "^M"
	case '\x00':
		return "^@"
	case '\x07':
		return "^G"
	case '\x08':
		return "^H"
	case '\x0c':
		return "^L"
	case '\x0b':
		return "^K"
	default:
		if ch < 32 {
			// Управляющие символы
			return fmt.Sprintf("^%c", rune(ch+'@'))
		}
		if ch == 127 {
			return "^?"
		}
		if ch >= 128 && ch < 160 {
			// Расширенные управляющие символы
			return fmt.Sprintf("^%c", rune((ch-128)+'@'))
		}
		return string(ch)
	}
}

// ProcessInput читает из файла или stdin и выводит с учётом флагов
func ProcessInput(file *os.File, flags CatFlags, lineNum *int) error {
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		isEmpty := strings.TrimSpace(line) == ""

		// Если используется -b
		if flags.NumberBlank {
			if isEmpty {
				// Пустая строка выводим без номера
				if flags.ShowAll {
					fmt.Println("$")
				} else {
					fmt.Println()
				}
			} else {
				// Непустая строка с номером
				if flags.ShowAll {
					result := fmt.Sprintf("%6d\t", *lineNum)
					for _, ch := range line {
						result += VisibleizeNonPrintable(ch)
					}
					result += "$"
					fmt.Println(result)
				} else {
					fmt.Printf("%6d\t%s\n", *lineNum, line)
				}
				*lineNum++
			}
		} else if flags.LineNumbers {
			// Если используется -n
			if flags.ShowAll {
				result := fmt.Sprintf("%6d\t", *lineNum)
				for _, ch := range line {
					result += VisibleizeNonPrintable(ch)
				}
				result += "$"
				fmt.Println(result)
			} else {
				fmt.Printf("%6d\t%s\n", *lineNum, line)
			}
			*lineNum++
		} else {
			// Без флагов нумерации
			if flags.ShowAll {
				result := ""
				for _, ch := range line {
					result += VisibleizeNonPrintable(ch)
				}
				result += "$"
				fmt.Println(result)
			} else {
				fmt.Println(line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// CatFile выводит содержимое одного файла
func CatFile(filename string, flags CatFlags, lineNum *int) error {
	var file *os.File
	var err error

	if filename == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(filename)
		if err != nil {
			return fmt.Errorf("ошибка: не удаётся открыть файл '%s': %v", filename, err)
		}
		defer file.Close()
	}

	return ProcessInput(file, flags, lineNum)
}

func main() {
	flags, files, err := ParseCatFlags(os.Args[1:])
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("Используй ./cat --help для справки")
		os.Exit(1)
	}

	// Если запрошена справка выводим её и выходим
	if flags.Help {
		ShowHelp()
		return
	}

	// Если файлы не указаны читаем из stdin
	if len(files) == 0 {
		files = []string{"-"}
	}

	// Выводим содержимое всех файлов
	lineNum := 1
	exitCode := 0
	for _, file := range files {
		err := CatFile(file, flags, &lineNum)
		if err != nil {
			fmt.Printf("%v\n", err)
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}
