package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BangFlags содержит флаги для управления командой lastcmd
type BangFlags struct {
	Help bool // -h/--help справка
}

// ShowHelp выводит справку по использованию команды lastcmd
func ShowHelp() {
	help := `
Использование: lastcmd [КЛЮЧИ] [аргументы]

Ключи:
  -h         Показать эту справку
  --help     Показать эту справку

Описание:
  lastcmd - выполняет последнюю команду из истории bash с опциональными аргументами.
  
  Команда читает последнюю выполненную команду из файла истории (~/.bash_history),
  выводит её на экран и выполняет.

Примеры:
  ./lastcmd             Выполнить последнюю команду как есть
`
	fmt.Println(help)
}

// ParseBangFlags парсит флаги для команды lastcmd
func ParseBangFlags(args []string) (BangFlags, []string, error) {
	flags := BangFlags{}
	var remainingArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			flags.Help = true
		} else if strings.HasPrefix(arg, "-") && arg != "-" {
			// Обработка комбинированных флагов
			for j := 1; j < len(arg); j++ {
				ch := rune(arg[j])
				switch ch {
				case 'h':
					flags.Help = true
				default:
					return flags, nil, fmt.Errorf("неизвестный флаг: -%c", ch)
				}
			}
		} else {
			remainingArgs = append(remainingArgs, arg)
		}
	}

	return flags, remainingArgs, nil
}

// GetHistoryFilePath получает путь к файлу истории bash
func GetHistoryFilePath() (string, error) {
	// Если задан HISTFILE используем его
	if histFile := os.Getenv("HISTFILE"); histFile != "" {
		return histFile, nil
	}

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("не удалось получить домашнюю директорию: %v", err)
		}
	}

	return filepath.Join(homeDir, ".bash_history"), nil
}

// / ReadHistoryFromFile читает все строки истории из файла
func ReadHistoryFromFile(histFile string) ([]string, error) {
	file, err := os.Open(histFile)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл истории '%s': %v", histFile, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Игнорируем пустые строки
		if line == "" {
			continue
		}

		// Игнорируем комментарии (timestamp)
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Игнорируем только точный вызов lastcmd
		if strings.HasPrefix(line, "lastcmd") ||
			strings.HasPrefix(line, "./lastcmd") ||
			strings.HasPrefix(line, "/home/user/Go/lastcmd/lastcmd") {
			continue
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении файла истории: %v", err)
	}

	return lines, nil
}

// GetLastCommand получает последнюю команду из истории
func GetLastCommand(histFile string) (string, error) {
	history, err := ReadHistoryFromFile(histFile)
	if err != nil {
		return "", err
	}

	if len(history) == 0 {
		return "", fmt.Errorf("история пуста")
	}

	return history[len(history)-1], nil
}

// BuildFinalCommand формирует итоговую команду
func BuildFinalCommand(cmdLine string, args []string) []string {
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return []string{}
	}

	if len(args) > 0 {
		// Берём только команду, заменяем аргументы
		return append([]string{parts[0]}, args...)
	}

	return parts
}

// ExecuteCommand выполняет команду через bash
func ExecuteCommand(finalArgs []string) error {
	if len(finalArgs) == 0 {
		return fmt.Errorf("нет команды для выполнения")
	}

	// Собираем команду в одну строку
	cmdLine := strings.Join(finalArgs, " ")
	cmd := exec.Command("bash", "-c", cmdLine)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("ошибка выполнения команды: %v", err)
	}

	return nil
}

func main() {
	flags, args, err := ParseBangFlags(os.Args[1:])
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("Используй lastcmd --help для справки")
		os.Exit(1)
	}

	if flags.Help {
		ShowHelp()
		return
	}

	histFile, err := GetHistoryFilePath()
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		os.Exit(1)
	}

	// Получаем последнюю команду
	cmdLine, err := GetLastCommand(histFile)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		os.Exit(1)
	}

	// Формируем итоговую команду
	finalArgs := BuildFinalCommand(cmdLine, args)
	if len(finalArgs) == 0 {
		fmt.Fprintln(os.Stderr, "Ошибка: не удалось построить команду")
		os.Exit(1)
	}
	fmt.Println(strings.Join(finalArgs, " "))
	err = ExecuteCommand(finalArgs)
	if err != nil {
		fmt.Printf("Ошибка выполнения команды: %v\n", err)
		os.Exit(1)
	}
}
