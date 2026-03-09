package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

// WcOptions содержит флаги для управления wc
type WcOptions struct {
	Help  bool
	Bytes bool // -c количество байт
	Chars bool // -m количество символов
	Lines bool // -l количество строк
	Files []string
}

// WcResult содержит результаты подсчёта для одного файла
type WcResult struct {
	Lines int    // Количество строк
	Words int    // Количество слов
	Bytes int    // Количество байт
	Chars int    // Количество символов
	File  string // Имя файла
}

// showHelp выводит справку по использованию команды
func showHelp() {
	fmt.Println(`Использование: wc [КЛЮЧИ] [ФАЙЛ]...

Ключи:
  -c           выводить количество байт
  -m           выводить количество символов
  -l           выводить количество строк
  -h, --help   показать эту справку

Описание:
  wc — команда для подсчёта количества строк, слов и байт в файлах. 
  Считает статистику текстовых файлов и выводит результаты в виде чисел.

Примеры:
  ./wc file.txt              Показать всё
  ./wc -l file.txt           Только строки
  ./wc -c file.txt           Только байты
  ./wc -m file.txt           Только символы`)
}

// parseFlags парсит аргументы командной строки
func parseFlags(args []string) (WcOptions, error) {
	opts := WcOptions{}
	files := []string{}

	// Проходим по всем аргументам
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") && arg != "-" {
			if arg == "--help" || arg == "-h" {
				opts.Help = true
			} else if strings.HasPrefix(arg, "--") {
				// Неизвестные длинные флаги
				return opts, fmt.Errorf("неизвестный флаг: %s", arg)
			} else {
				// комбинированные флаги
				for _, ch := range arg[1:] {
					switch ch {
					case 'l':
						opts.Lines = true
					case 'c':
						opts.Bytes = true
					case 'm':
						opts.Chars = true
					case 'h':
						opts.Help = true
					default:
						return opts, fmt.Errorf("неизвестный флаг: -%c", ch)
					}
				}
			}
		} else {
			files = append(files, arg)
		}
	}

	opts.Files = files
	return opts, nil
}

// processFile обрабатывает один файл и подсчитывает статистику
func processFile(filename string) (*WcResult, error) {
	result := &WcResult{File: filename}

	var file *os.File
	var err error

	// Открываем файл или stdin
	if filename == "-" || filename == "" {
		file = os.Stdin
		result.File = ""
	} else {
		file, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()
	}

	// Читаем файл построчно
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		result.Lines++
		result.Words += len(strings.Fields(line))
		// Байты включают символ новой строки
		result.Bytes += len([]byte(line)) + 1
		// Символы включают символ новой строки
		result.Chars += utf8.RuneCountInString(line) + 1
	}

	return result, scanner.Err()
}

// printResult выводит результаты в правильном формате
func printResult(r *WcResult, opts WcOptions) {
	// Если никаких флагов не указано выводим всё
	if !opts.Lines && !opts.Bytes && !opts.Chars {
		fmt.Printf("%d %d %d", r.Lines, r.Words, r.Bytes)
	} else {
		// Выводим только выбранные опции
		if opts.Lines {
			fmt.Printf("%d", r.Lines)
		}
		if opts.Chars {
			fmt.Printf("%d", r.Chars)
		}
		if opts.Bytes {
			fmt.Printf("%d", r.Bytes)
		}
	}

	// Добавляем имя файла если оно есть
	if r.File != "" {
		fmt.Printf(" %s", r.File)
	}
	fmt.Println()
}

func main() {
	opts, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "wc: %v\n", err)
		os.Exit(1)
	}

	if opts.Help {
		showHelp()
		return
	}

	// Если файлы не указаны читаем из stdin
	if len(opts.Files) == 0 {
		opts.Files = []string{"-"}
	}

	// Обрабатываем каждый файл
	var results []*WcResult
	for _, f := range opts.Files {
		r, err := processFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "wc: %s: %v\n", f, err)
			continue
		}
		results = append(results, r)
		printResult(r, opts)
	}

	if len(results) > 1 {
		total := &WcResult{File: "total"}
		for _, r := range results {
			total.Lines += r.Lines
			total.Words += r.Words
			total.Bytes += r.Bytes
			total.Chars += r.Chars
		}
		printResult(total, opts)
	}
}
