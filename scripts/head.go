package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// HeadOptions содержит флаги для управления head
type HeadOptions struct {
	Help  bool
	Bytes int
	Lines int
	Quiet bool
	Files []string
}

// Функция для вывода справки
func printHelpHead() {
	helpText := `Использование: head [КЛЮЧИ] [ФАЙЛ]...

Ключи:
  -c K             Напечатать первые K байт каждого файла
  -n K             Напечатать K строк каждого файла (по умолчанию 10)
  -q               Не печатать заголовки с именами файлов
  -h      		   Показать эту справку
  --help		   Показать эту справку

Примеры:
  ./head -n 5 file.txt        Показать первые 5 строк
  ./head -c 100 file.txt      Показать первые 100 байт
  ./head -qn 5 file.txt       Комбинированные флаги + вывод без заголовка`
	fmt.Println(helpText)
}

// Вспомогательная функция для парсинга флага с значением
func parseFlag(args []string, i int, flagChar rune) (int, int, error) {
	if i+1 >= len(args) {
		return 0, i, fmt.Errorf("ключ должен использоваться с аргументом — «%c»", flagChar)
	}
	num, err := strconv.Atoi(strings.TrimSpace(args[i+1]))
	return num, i + 2, err
}

// Функция для парсинга аргументов вручную
func parseArgumentsHead() (*HeadOptions, error) {
	options := &HeadOptions{
		Lines: 10,
		Bytes: -1,
	}

	args := os.Args[1:]
	i := 0

	for i < len(args) {
		arg := args[i]

		if arg == "-h" || arg == "--help" {
			options.Help = true
			i++
		} else if arg == "-q" {
			options.Quiet = true
			i++
		} else if arg == "-c" {
			num, newI, err := parseFlag(args, i, 'c')
			if err != nil {
				return nil, err
			}
			options.Bytes = num
			i = newI
		} else if arg == "-n" {
			num, newI, err := parseFlag(args, i, 'n')
			if err != nil {
				return nil, err
			}
			options.Lines = num
			i = newI
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Комбинированные флаги
			flags := arg[1:]
			j := 0

			for j < len(flags) {
				switch flags[j] {
				case 'q':
					options.Quiet = true
					j++
				case 'h':
					options.Help = true
					j++
				case 'c':
					if j+1 < len(flags) {
						numStr := flags[j+1:]
						num, err := strconv.Atoi(numStr)
						if err != nil {
							return nil, fmt.Errorf("неверное значение для -c: %s", numStr)
						}
						options.Bytes = num
						j = len(flags)
					} else if i+1 < len(args) {
						i++
						num, err := strconv.Atoi(strings.TrimSpace(args[i]))
						if err != nil {
							return nil, fmt.Errorf("неверное значение для -c: %s", args[i])
						}
						options.Bytes = num
						j = len(flags)
					} else {
						return nil, fmt.Errorf("ключ должен использоваться с аргументом — «c»")
					}
				case 'n':
					if j+1 < len(flags) {
						numStr := flags[j+1:]
						num, err := strconv.Atoi(numStr)
						if err != nil {
							return nil, fmt.Errorf("неверное значение для -n: %s", numStr)
						}
						options.Lines = num
						j = len(flags)
					} else if i+1 < len(args) {
						i++
						num, err := strconv.Atoi(strings.TrimSpace(args[i]))
						if err != nil {
							return nil, fmt.Errorf("неверное значение для -n: %s", args[i])
						}
						options.Lines = num
						j = len(flags)
					} else {
						return nil, fmt.Errorf("ключ должен использоваться с аргументом — «n»")
					}
				default:
					return nil, fmt.Errorf("неизвестный флаг: -%c", flags[j])
				}
			}
			i++
		} else {
			options.Files = append(options.Files, arg)
			i++
		}
	}

	return options, nil
}

// Вспомогательная функция для вывода строк из файла
func headLinesFromFile(file *os.File, lines int, printHeader bool) error {
	if printHeader {
		fmt.Printf("==> %s <==\n", file.Name())
	}

	if lines < 0 {
		// Читаем все строки и выводим все кроме последних N
		var allLines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			allLines = append(allLines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return err
		}

		end := len(allLines) + lines
		if end < 0 {
			end = 0
		}
		for i := 0; i < end; i++ {
			fmt.Println(allLines[i])
		}
	} else {
		// Выводим первые N строк
		scanner := bufio.NewScanner(file)
		count := 0
		for scanner.Scan() && count < lines {
			fmt.Println(scanner.Text())
			count++
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}
	return nil
}

// Вспомогательная функция для вывода байт из файла
func headBytesFromFile(file *os.File, bytes int, printHeader bool) error {
	if printHeader {
		fmt.Printf("==> %s <==\n", file.Name())
	}

	data, err := os.ReadFile(file.Name())
	if err != nil {
		return err
	}

	if bytes < 0 {
		end := len(data) + bytes
		if end < 0 {
			end = 0
		}
		fmt.Print(string(data[:end]))
	} else {
		if bytes > len(data) {
			bytes = len(data)
		}
		fmt.Print(string(data[:bytes]))
	}
	return nil
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "head: критическая ошибка (panic): %v\n", r)
			os.Exit(1)
		}
	}()

	options, err := parseArgumentsHead()
	if err != nil {
		fmt.Fprintf(os.Stderr, "head: %v\n", err)
		fmt.Fprintf(os.Stderr, "По команде «head --help» можно получить дополнительную информацию.\n")
		os.Exit(1)
	}

	if options.Help {
		printHelpHead()
		os.Exit(0)
	}

	if len(options.Files) == 0 {
		// Читаем из stdin
		if options.Bytes != -1 {
			headBytesFromFile(os.Stdin, options.Bytes, false)
		} else {
			headLinesFromFile(os.Stdin, options.Lines, false)
		}
	} else {
		for i, filename := range options.Files {
			printHeader := len(options.Files) > 1 && !options.Quiet

			file, err := os.Open(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "head: не могу открыть файл %q: %v\n", filename, err)
				continue
			}

			if options.Bytes != -1 {
				headBytesFromFile(file, options.Bytes, printHeader)
			} else {
				headLinesFromFile(file, options.Lines, printHeader)
			}

			file.Close()

			if i < len(options.Files)-1 && printHeader {
				fmt.Println()
			}
		}
	}

	os.Exit(0)
}
