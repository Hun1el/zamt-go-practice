package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// TailOptions содержит флаги для управления tail
type TailOptions struct {
	Help        bool
	Bytes       int
	Lines       int
	BytesOffset bool
	LinesOffset bool
	Quiet       bool
	Files       []string
}

// Функция для вывода справки
func printHelpTail() {
	helpText := `Использование: tail [КЛЮЧИ] [ФАЙЛ]...
  
Ключи:
  -c K             Напечатать последние K байт каждого файла
                   Если перед K стоит «+», напечатать начиная с K байта
  -n K             Напечатать последние K строк каждого файла (по умолчанию 10)
                   Если перед K стоит «+», напечатать начиная со строки K
  -q               Не печатать заголовки с именами файлов
  -h     		   Показать эту справку
  --help  		   Показать эту справку

Описание:
  tail - выводит последние строки или байты из файлов.
            
Примеры:
  ./tail -n 5 file.txt       		Показать последние 5 строк
  ./tail -c 100 file.txt     		Показать последние 100 байт
  ./tail -n +5 file.txt      		Показать начиная со строки 5
  ./tail -c +100 file.txt    		Показать начиная с байта 100
  ./tail -q file.txt file.txt   	Без заголовков файлов`
	fmt.Println(helpText)
}

// Функция для парсинга числа с возможным плюсом в начале
func parseNumber(s string) (int, bool, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "+") {
		num, err := strconv.Atoi(s[1:])
		return num, true, err
	}
	num, err := strconv.Atoi(s)
	return num, false, err
}

// Вспомогательная функция для парсинга флага с значением
func parseFlag(args []string, i int, flagChar rune) (int, bool, int, error) {
	if i+1 >= len(args) {
		return 0, false, i, fmt.Errorf("ключ должен использоваться с аргументом — «%c»", flagChar)
	}
	num, isOffset, err := parseNumber(args[i+1])
	return num, isOffset, i + 2, err
}

// Функция для парсинга аргументов вручную
func parseArgumentsTail() (*TailOptions, error) {
	options := &TailOptions{
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
			num, isOffset, newI, err := parseFlag(args, i, 'c')
			if err != nil {
				return nil, err
			}
			options.Bytes = num
			options.BytesOffset = isOffset
			i = newI
		} else if arg == "-n" {
			num, isOffset, newI, err := parseFlag(args, i, 'n')
			if err != nil {
				return nil, err
			}
			options.Lines = num
			options.LinesOffset = isOffset
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
						num, isOffset, err := parseNumber(numStr)
						if err != nil {
							return nil, fmt.Errorf("неверное значение для -c: %s", numStr)
						}
						options.Bytes = num
						options.BytesOffset = isOffset
						j = len(flags)
					} else if i+1 < len(args) {
						i++
						num, isOffset, err := parseNumber(args[i])
						if err != nil {
							return nil, fmt.Errorf("неверное значение для -c: %s", args[i])
						}
						options.Bytes = num
						options.BytesOffset = isOffset
						j = len(flags)
					} else {
						return nil, fmt.Errorf("ключ должен использоваться с аргументом — «c»")
					}
				case 'n':
					if j+1 < len(flags) {
						numStr := flags[j+1:]
						num, isOffset, err := parseNumber(numStr)
						if err != nil {
							return nil, fmt.Errorf("неверное значение для -n: %s", numStr)
						}
						options.Lines = num
						options.LinesOffset = isOffset
						j = len(flags)
					} else if i+1 < len(args) {
						i++
						num, isOffset, err := parseNumber(args[i])
						if err != nil {
							return nil, fmt.Errorf("неверное значение для -n: %s", args[i])
						}
						options.Lines = num
						options.LinesOffset = isOffset
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

// Вспомогательная функция для вывода данных по строкам/байтам
func outputLines(allLines []string, lines int, isOffset bool, printHeader bool, filename string) {
	if printHeader {
		fmt.Printf("==> %s <==\n", filename)
	}

	if isOffset {
		startIdx := lines - 1
		if startIdx < len(allLines) && startIdx >= 0 {
			for i := startIdx; i < len(allLines); i++ {
				fmt.Println(allLines[i])
			}
		}
	} else {
		start := len(allLines) - lines
		if start < 0 {
			start = 0
		}
		for i := start; i < len(allLines); i++ {
			fmt.Println(allLines[i])
		}
	}
}

func outputBytes(data []byte, bytes int, isOffset bool, printHeader bool, filename string) {
	if printHeader {
		fmt.Printf("==> %s <==\n", filename)
	}

	if isOffset {
		if bytes < len(data) {
			fmt.Print(string(data[bytes:]))
		}
	} else {
		if len(data) > bytes {
			fmt.Print(string(data[len(data)-bytes:]))
		} else {
			fmt.Print(string(data))
		}
	}
}

// Функция для обработки файла
func tailFile(filename string, options *TailOptions, printHeader bool) error {
	if options.Bytes != -1 {
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tail: не могу прочитать файл %q: %v\n", filename, err)
			return err
		}
		outputBytes(data, options.Bytes, options.BytesOffset, printHeader, filename)
	} else {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tail: не могу открыть файл %q: %v\n", filename, err)
			return err
		}
		defer file.Close()

		var allLines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			allLines = append(allLines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "tail: ошибка чтения: %v\n", err)
			return err
		}

		outputLines(allLines, options.Lines, options.LinesOffset, printHeader, filename)
	}
	return nil
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "tail: критическая ошибка (panic): %v\n", r)
			os.Exit(1)
		}
	}()

	options, err := parseArgumentsTail()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tail: %v\n", err)
		fmt.Fprintf(os.Stderr, "По команде «tail --help» можно получить дополнительную информацию.\n")
		os.Exit(1)
	}

	if options.Help {
		printHelpTail()
		os.Exit(0)
	}

	if len(options.Files) == 0 {
		options.Files = append(options.Files, "-")
	}

	for i, filename := range options.Files {
		printHeader := len(options.Files) > 1 && !options.Quiet

		if filename == "-" {
			tailFile("/dev/stdin", options, printHeader)
		} else {
			tailFile(filename, options, printHeader)
		}

		if i < len(options.Files)-1 && printHeader {
			fmt.Println()
		}
	}

	os.Exit(0)
}
