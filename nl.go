package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// NLFlags содержит флаги для управления nl
type NLFlags struct {
	Help             bool   // --help справка
	BodyNumbering    string // -b стиль нумерации тела
	HeaderNumbering  string // -h стиль нумерации заголовка
	FooterNumbering  string // -f стиль нумерации нижнего колонтитула
	SectionDelimiter string // -d разделитель секций
	Files            []string
}

// showHelp выводит справку по использованию команды nl
func showHelp() {
	fmt.Println(`Использование: nl [КЛЮЧИ] [ФАЙЛ]...

Ключи:
  -b СТИЛЬ    использовать СТИЛЬ нумерации строк тела
              (t - только непустые, a - все, n - нет)
  -d СС       использовать СС как логический разделитель страниц
  -f СТИЛЬ    использовать СТИЛЬ нумерации строк нижнего колонтитула
  -h СТИЛЬ    использовать СТИЛЬ нумерации строк верхнего колонтитула
  --help      показать эту справку

Описание:
  nl - выводит файл с номерами строк.

Примеры:
  ./nl file.txt                    Нумеровать файл (только непустые)
  ./nl -b a file.txt               Нумеровать все строки, включая пустые
  ./nl -b n file.txt               Не нумеровать (только вывести строки)
  ./nl -h a file.txt               Нумеровать заголовок со всеми строками
  ./nl -f t file.txt               Нумеровать нижний колонтитул только непустые
  ./nl -d '::' file.txt            Использовать '::' как разделитель секций`)
}

// parseFlags парсит флаги для nl
func parseFlags(args []string) (NLFlags, error) {
	flags := NLFlags{
		BodyNumbering:    "t",
		HeaderNumbering:  "n",
		FooterNumbering:  "n",
		SectionDelimiter: "\\:",
	}
	files := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") && arg != "-" {
			switch arg {
			case "--help":
				flags.Help = true
			case "-b":
				if i+1 < len(args) {
					i++
					flags.BodyNumbering = args[i]
				} else {
					return flags, fmt.Errorf("флаг -b требует аргумента")
				}
			case "-h":
				if i+1 < len(args) {
					i++
					flags.HeaderNumbering = args[i]
				} else {
					return flags, fmt.Errorf("флаг -h требует аргумента")
				}
			case "-d":
				if i+1 < len(args) {
					i++
					flags.SectionDelimiter = args[i]
				} else {
					return flags, fmt.Errorf("флаг -d требует аргумента")
				}
			case "-f":
				if i+1 < len(args) {
					i++
					flags.FooterNumbering = args[i]
				} else {
					return flags, fmt.Errorf("флаг -f требует аргумента")
				}
			default:
				return flags, fmt.Errorf("неизвестный флаг: %s", arg)
			}
		} else {
			files = append(files, arg)
		}
	}

	flags.Files = files
	return flags, nil
}

// shouldNumber определяет должна ли строка быть пронумерована
func shouldNumber(line string, style string) bool {
	switch style {
	case "t":
		// Нумеровать только непустые строки
		return strings.TrimSpace(line) != ""
	case "a":
		// Нумеровать все строки, включая пустые
		return true
	case "n":
		// Не нумеровать
		return false
	default:
		// По умолчанию как t
		return strings.TrimSpace(line) != ""
	}
}

// processFile обрабатывает файл
func processFile(filename string, flags NLFlags) error {
	var file *os.File
	var err error

	if filename == "" || filename == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	scanner := bufio.NewScanner(file)
	bodyNum := 1
	footerNum := 1
	currentStyle := flags.BodyNumbering
	currentNum := &bodyNum

	for scanner.Scan() {
		line := scanner.Text()

		// Проверяем разделитель секций
		if strings.HasPrefix(line, flags.SectionDelimiter) {
			// Выводим разделитель
			fmt.Println()
			// Сбрасываем счётчик нумерации после разделителя
			bodyNum = 1
			currentStyle = flags.FooterNumbering
			currentNum = &footerNum
			continue
		}

		// Определяем стиль нумерации для текущей строки
		style := currentStyle

		// Проверяем нужна ли нумерация для этой строки
		if shouldNumber(line, style) {
			fmt.Printf("%6d\t%s\n", *currentNum, line)
			(*currentNum)++
		} else {
			// Если не нумеруем, но это стиль -b a или -d с непустой строкой
			// Увеличиваем счётчик для непустых строк в режиме -b a
			if style == "a" && strings.TrimSpace(line) == "" {
				// Пустая строка при -b a выводится без номера
				fmt.Printf("      \t%s\n", line)
			} else if style == "a" {
				// Непустая строка при -b a
				fmt.Printf("%6d\t%s\n", *currentNum, line)
				(*currentNum)++
			} else if style == "n" {
				// Без нумерации
				fmt.Printf("      \t%s\n", line)
			} else {
				// -b t (только непустые)пустая строка
				fmt.Printf("      \t%s\n", line)
			}
		}
	}

	return scanner.Err()
}

func main() {
	flags, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "nl: %v\n", err)
		os.Exit(1)
	}

	if flags.Help {
		showHelp()
		return
	}

	// Если файлы не указаны читаем из stdin
	if len(flags.Files) == 0 {
		err := processFile("-", flags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nl: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Обрабатываем каждый файл
	for _, file := range flags.Files {
		err := processFile(file, flags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nl: не могу открыть '%s': %v\n", file, err)
			os.Exit(1)
		}
	}
}
