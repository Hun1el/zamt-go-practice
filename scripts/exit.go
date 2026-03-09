package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// ExitFlags содержит флаги для управления exit
type ExitFlags struct {
	Help bool // -h/--help справка
}

// showHelp выводит справку по использованию команды exit
func showHelp() {
	fmt.Println(`Использование: exit [n]

Ключи:
  -h         показать эту справку
  --help     показать эту справку

Описание:
  exit - завершает текущий сеанс оболочки с указанным кодом выхода.
  
  n          код выхода (по умолчанию последний код команды или 0).


Примеры:
  ./exit         Выход с кодом 0
  ./exit 0       Выход с кодом 0 (успех)
  ./exit 1       Выход с кодом 1 (ошибка)`)
}

// parseFlags парсит флаги для exit
func parseFlags(args []string) (ExitFlags, string, error) {
	flags := ExitFlags{}
	var code string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") && arg != "-" {
			if _, err := strconv.Atoi(arg); err == nil {
				code = arg
				continue
			}

			if arg == "--help" || arg == "-h" {
				flags.Help = true
			} else if len(arg) > 2 {
				for _, ch := range arg[1:] {
					switch ch {
					case 'h':
						flags.Help = true
					default:
						return flags, "", fmt.Errorf("неизвестный флаг: %s", arg)
					}
				}
			} else {
				switch arg {
				case "-h":
					flags.Help = true
				default:
					return flags, "", fmt.Errorf("неизвестный флаг: %s", arg)
				}
			}
		} else {
			code = arg
		}
	}

	return flags, code, nil
}

// closeParentShell отправляет SIGHUP родительскому процессу
func closeParentShell(exitCode int) {
	ppid := os.Getppid()

	if ppid > 1 {
		parentProcess, err := os.FindProcess(ppid)
		if err == nil {
			parentProcess.Signal(syscall.SIGHUP)
		}
	}
}

func main() {
	flags, codeStr, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "exit: %v\n", err)
		os.Exit(1)
	}

	if flags.Help {
		showHelp()
		os.Exit(0)
	}

	exitCode := 0

	// Если передан код выхода
	if codeStr != "" {
		code, err := strconv.Atoi(codeStr)
		if err != nil {
			// При ошибке парсинга выходим с кодом 1
			exitCode = 1
		} else {
			// Берём остаток от деления на 256
			exitCode = code % 256
			// Если отрицательное число добавляем 256
			if exitCode < 0 {
				exitCode += 256
			}
		}
	}

	// Закрыть родительскую оболочку
	closeParentShell(exitCode)

	// Выход с кодом
	os.Exit(exitCode)
}
