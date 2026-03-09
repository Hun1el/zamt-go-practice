package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PwdFlags содержит флаги для управления выводом pwd
type PwdFlags struct {
	Logical  bool // -L логический путь
	Physical bool // -P физический путь
	Help     bool // --help справка
}

// ShowHelp выводит справку по использованию pwd
func ShowHelp() {
	help := `
Использование: ./pwd [КЛЮЧИ]

Ключи:
  -L         Логический путь (показывает пути со следованием символическим ссылкам)
  -P         Физический путь (показывает реальный путь без символических ссылок)
  -h         Показать эту справку
  --help     Показать эту справку

Описание:
  pwd выводит абсолютный путь текущей рабочей директории.

Примеры:
  ./pwd              Вывести текущую директорию
  ./pwd -L           Вывести логический путь со ссылками
  ./pwd -P           Вывести физический путь без ссылок
  ./pwd --help       Показать эту справку
`
	fmt.Println(help)
}

// ParsePwdFlags парсит флаги для pwd
func ParsePwdFlags(args []string) (PwdFlags, error) {
	flags := PwdFlags{}

	// По умолчанию используем логический путь
	flags.Logical = true

	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			flags.Help = true
		} else if arg == "-L" {
			flags.Logical = true
			flags.Physical = false
		} else if arg == "-P" {
			flags.Physical = true
			flags.Logical = false
		} else if strings.HasPrefix(arg, "-") {
			return flags, fmt.Errorf("неизвестный флаг: %s", arg)
		}
	}

	return flags, nil
}

// GetCurrentDirectory получает текущую директорию
func GetCurrentDirectory(logical bool) (string, error) {
	if logical {
		// Логический путь используем переменную окружения PWD если она установлена
		pwdEnv := os.Getenv("PWD")
		if pwdEnv != "" {
			return pwdEnv, nil
		}
	}

	// Физический путь или если PWD не установлена
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("ошибка при получении текущей директории: %v", err)
	}

	if logical {
		return workDir, nil
	}

	// Для физического пути разрешаем все символические ссылки
	realPath, err := filepath.EvalSymlinks(workDir)
	if err != nil {
		return "", fmt.Errorf("ошибка при разрешении символических ссылок: %v", err)
	}

	return realPath, nil
}

// ValidateDirectory проверяет существование и доступность директории
func ValidateDirectory(path string) error {
	// Проверяем существование директории
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("ошибка: не удаётся получить доступ к директории '%s': %v", path, err)
	}

	// Проверяем что это директория
	if !info.IsDir() {
		return fmt.Errorf("ошибка: '%s' не является директорией", path)
	}

	return nil
}

func main() {
	flags, err := ParsePwdFlags(os.Args[1:])
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("Используй ./pwd --help для справки")
		os.Exit(1)
	}

	if flags.Help {
		ShowHelp()
		return
	}

	// Получаем текущую директорию
	currentDir, err := GetCurrentDirectory(flags.Logical)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		os.Exit(1)
	}

	// Проверяем существование директории
	err = ValidateDirectory(currentDir)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	// Выводим путь текущей директории
	fmt.Println(currentDir)
}
