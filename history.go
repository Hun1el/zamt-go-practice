package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// HistoryFlags содержит флаги для управления history
type HistoryFlags struct {
	Write bool   // -w записать историю в файл
	Clear bool   // -c очистить историю
	Read  bool   // -r прочитать файл истории и добавить
	Help  bool   // -h/--help справка
	File  string // имя файла для операций
}

// showHelp выводит справку по использованию команды history
func showHelp() {
	fmt.Println(`Использование: history [КЛЮЧИ] [n]

Ключи:
  -r         прочитать файл истории
  -w file    записать всю историю в файл file
  -c         очистить файл истории
  -h         показать эту справку
  --help     показать эту справку

Описание:
  history - выводит список истории выполненных команд.
  
  -r         читает файл истории и добавляет строки в текущий список
  -w         записывает всю историю
  -c         очищает файл истории

Примеры:
  ./history              Показать историю команд
  ./history 10           Показать последние 10 команд
  ./history -r           Прочитать файл истории и показать
  ./history -w file      Записать всю историю в файл
  ./history -c           Очистить историю`)
}

// parseFlags парсит флаги для history
func parseFlags(args []string) (HistoryFlags, string, error) {
	flags := HistoryFlags{}
	var count string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") && arg != "-" {
			if arg == "--help" || arg == "-h" {
				flags.Help = true
			} else if len(arg) > 2 {
				for _, ch := range arg[1:] {
					switch ch {
					case 'w':
						flags.Write = true
					case 'c':
						flags.Clear = true
					case 'r':
						flags.Read = true
					case 'h':
						flags.Help = true
					default:
						return flags, "", fmt.Errorf("неизвестный флаг: %s", arg)
					}
				}
			} else {
				switch arg {
				case "-w":
					flags.Write = true
				case "-c":
					flags.Clear = true
				case "-r":
					flags.Read = true
				case "-":
					return flags, "", fmt.Errorf("неизвестный флаг: %s", arg)
				default:
					return flags, "", fmt.Errorf("неизвестный флаг: %s", arg)
				}
			}
		} else {
			// Если есть флаг -w следующий аргумент это имя файла
			if flags.Write && flags.File == "" {
				flags.File = arg
			} else {
				count = arg
			}
		}
	}

	return flags, count, nil
}

// getHistoryFile возвращает путь к файлу истории bash
func getHistoryFile() string {
	histfile := os.Getenv("HISTFILE")
	if histfile != "" {
		return histfile
	}

	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".bash_history")
}

// readHistory читает историю из файла
func readHistory(filename string) []string {
	var history []string

	file, err := os.Open(filename)
	if err != nil {
		return history
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			history = append(history, line)
		}
	}

	return history
}

// writeHistory записывает историю в файл
func writeHistory(filename string, history []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, cmd := range history {
		fmt.Fprintln(file, cmd)
	}

	return nil
}

// appendHistory добавляет строку в конец файла истории
func appendHistory(filename string, line string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, line)
	return err
}

// displayHistory выводит историю команд
func displayHistory(history []string, count string) {
	if len(history) == 0 {
		fmt.Println("История команд пуста")
		return
	}

	start := 0
	if count != "" {
		n, err := strconv.Atoi(count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "history: %s: не число\n", count)
			fmt.Fprintln(os.Stderr, "Используй history --help для справки")
			os.Exit(1)
		}

		if n > 0 && n < len(history) {
			start = len(history) - n
		}
	}

	for i := start; i < len(history); i++ {
		fmt.Printf("%4d  %s\n", i+1, history[i])
	}
}

func main() {
	flags, count, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "history: %v\n", err)
		fmt.Fprintln(os.Stderr, "Используй history --help для справки")
		os.Exit(1)
	}

	if flags.Help {
		showHelp()
		return
	}

	historyFile := getHistoryFile()

	// Очистить историю
	if flags.Clear {
		writeHistory(historyFile, []string{})
		fmt.Printf("История очищена\n")
		return
	}

	// Записать историю
	if flags.Write {
		targetFile := historyFile
		if flags.File != "" {
			targetFile = flags.File
		}
		history := readHistory(historyFile)
		writeHistory(targetFile, history)
		fmt.Printf("История записана в: %s\n", targetFile)
		return
	}

	// Прочитать файл истории и показать
	if flags.Read {
		history := readHistory(historyFile)
		fmt.Printf("Загружена история из: %s\n", historyFile)
		displayHistory(history, count)
		return
	}

	// Показать историю
	history := readHistory(historyFile)
	displayHistory(history, count)
}
