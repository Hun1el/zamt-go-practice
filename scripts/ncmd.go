package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ShowHelp выводит справку по использованию команды ncmd
func ShowHelp() {
	help := `
Использование: ncmd N [аргументы]

Ключи:
  -h         Показать эту справку
  --help     Показать эту справку

Аргументы:
  N          Номер команды из истории (считая с 1)

Описание:
  ncmd - выполняет N-ую команду из истории bash.
  
  Команда читает историю из файла (~/.bash_history) и выполняет
  команду под номером N.

Примеры:
  ./ncmd 1              Выполнить первую команду из истории
  ./ncmd 10             Выполнить десятую команду из истории
`
	fmt.Println(help)
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		ShowHelp()
		return
	}

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		fmt.Fprintln(os.Stderr, "ncmd: не задана переменная HOME")
		os.Exit(1)
	}
	histFile := filepath.Join(homeDir, ".bash_history")

	// Читаем историю из файла
	history := readHistoryFromFile(histFile)
	if len(history) == 0 {
		fmt.Fprintln(os.Stderr, "ncmd: история пуста")
		os.Exit(1)
	}

	// Должен быть хотя бы один аргумент
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "ncmd: требуется номер команды")
		fmt.Fprintln(os.Stderr, "Использование: ncmd N [аргументы...]")
		fmt.Fprintln(os.Stderr, "Используй ncmd --help для справки")
		os.Exit(1)
	}

	// Парсим номер
	n, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ncmd: '%s' не число\n", args[0])
		os.Exit(1)
	}
	if n < 1 || n > len(history) {
		fmt.Fprintf(os.Stderr, "ncmd: команда %d не найдена (есть 1–%d)\n", n, len(history))
		os.Exit(1)
	}

	// Берём команду по номеру
	cmdLine := history[n-1]

	// Разбиваем команду на части
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return
	}

	// Формируем итоговую команду
	var finalArgs []string
	if len(args) > 1 {
		finalArgs = append([]string{parts[0]}, args[1:]...)
	} else {
		finalArgs = parts
	}

	fmt.Printf("%s\n", strings.Join(finalArgs, " "))

	// Выполняем через bash для встроенных команд
	c := exec.Command("bash", "-c", strings.Join(finalArgs, " "))
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		os.Exit(1)
	}
}

func readHistoryFromFile(histFile string) []string {
	file, err := os.Open(histFile)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Игнорируем пустые строки и комментарии
		if line != "" && !strings.HasPrefix(line, "#") {
			// Игнорируем вызовы ncmd
			if !strings.HasPrefix(line, "ncmd ") &&
				!strings.HasPrefix(line, "./ncmd ") &&
				line != "ncmd" &&
				line != "./ncmd" {
				lines = append(lines, line)
			}
		}
	}
	return lines
}
