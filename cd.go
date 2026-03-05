package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CdFlags содержит флаги для управления выводом cd
type CdFlags struct {
	Logical  bool // -L логический путь
	Physical bool // -P физический путь
	NoError  bool // -e возвращать ошибку если директория не существует
	Help     bool // --help/-h справка
}

// ShowHelp выводит справку по использованию команды cd
func ShowHelp() {
	help := `
Использование: ./cd [КЛЮЧИ] [путь]

Ключи:
  -L         Логический путь (следует символическим ссылкам)
  -P         Физический путь (разрешает все символические ссылки)
  -e         Возвращать ошибку если директория не существует
  -h         Показать эту справку
  --help     Показать эту справку

Описание:
  cd - изменяет текущую рабочую директорию.

Специальные пути:
  .          Текущая директория
  ..         Родительская директория
  ~          Домашняя директория пользователя
  -          Предыдущая директория

Примеры:
  ./cd                     Перейти в домашнюю директорию (~)
  ./cd /home               Перейти в директорию /home (абсолютный путь)
  ./cd Doc                 Перейти в папку Documents (относительный путь)
  ./cd -L /path/to/link    Перейти по ссылке логически
  ./cd -P /path/to/link    Перейти по ссылке физически
  ./cd -e /none            Вернуть ошибку если директория не существует
  ./cd -Le /path           Логический путь и проверка существования
`
	fmt.Println(help)
}

// ParseCdFlags парсит флаги для cd
func ParseCdFlags(args []string) (CdFlags, string, error) {
	flags := CdFlags{}
	path := ""

	// По умолчанию используем логический путь
	flags.Logical = true

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			flags.Help = true
		} else if strings.HasPrefix(arg, "-") && arg != "-" {
			// Обработка комбинированных флагов
			for j := 1; j < len(arg); j++ {
				ch := rune(arg[j])
				switch ch {
				case 'L':
					flags.Logical = true
					flags.Physical = false
				case 'P':
					flags.Physical = true
					flags.Logical = false
				case 'e':
					flags.NoError = true
				case 'h':
					flags.Help = true
				default:
					return flags, "", fmt.Errorf("неизвестный флаг: -%c", ch)
				}
			}
		} else if arg == "-" {
			path = "-"
		} else {
			path = arg
		}
	}

	return flags, path, nil
}

// ExpandPath раскрывает специальные пути
func ExpandPath(pathStr string) (string, error) {
	// Обработка предыдущей директории
	if pathStr == "-" {
		oldPwd := os.Getenv("OLDPWD")
		if oldPwd == "" {
			return "", fmt.Errorf("ошибка: OLDPWD не установлена (предыдущая директория неизвестна)")
		}
		return oldPwd, nil
	}

	// Обработка домашней директории
	if strings.HasPrefix(pathStr, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("ошибка при получении домашней директории: %v", err)
		}

		if pathStr == "~" {
			return homeDir, nil
		}

		if strings.HasPrefix(pathStr, "~/") {
			return filepath.Join(homeDir, pathStr[2:]), nil
		}
	}

	return pathStr, nil
}

// ValidateDirectory проверяет существование и доступность директории
func ValidateDirectory(path string) error {
	// Проверяем существование директории
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("ошибка: не удаётся получить доступ к '%s': %v", path, err)
	}

	// Проверяем что это директория
	if !info.IsDir() {
		return fmt.Errorf("ошибка: '%s' не является директорией", path)
	}

	return nil
}

// GetAbsolutePath получает абсолютный путь
func GetAbsolutePath(pathStr string, logical bool) (string, error) {
	// Раскрываем специальные пути
	expandedPath, err := ExpandPath(pathStr)
	if err != nil {
		return "", err
	}

	var absolutePath string

	// Получаем абсолютный путь
	if filepath.IsAbs(expandedPath) {
		absolutePath = expandedPath
	} else {
		// Это относительный путь
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("ошибка при получении текущей директории: %v", err)
		}
		absolutePath = filepath.Join(currentDir, expandedPath)
	}

	absolutePath = filepath.Clean(absolutePath)

	// Если нужен физический путь разрешаем символические ссылки
	if !logical {
		realPath, err := filepath.EvalSymlinks(absolutePath)
		if err != nil {
			return "", fmt.Errorf("ошибка при разрешении символических ссылок: %v", err)
		}
		absolutePath = realPath
	}

	return absolutePath, nil
}

func main() {
	flags, path, err := ParseCdFlags(os.Args[1:])
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("Используй ./cd --help для справки")
		os.Exit(1)
	}

	if flags.Help {
		ShowHelp()
		return
	}

	// Получаем текущую директорию для OLDPWD
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = ""
	}

	// Если путь не указан переходим в домашнюю директорию
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Ошибка: не удаётся получить домашнюю директорию: %v\n", err)
			os.Exit(1)
		}
		path = homeDir
	}

	// Получаем абсолютный путь
	targetDir, err := GetAbsolutePath(path, flags.Logical)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		if flags.NoError {
			os.Exit(1)
		}
		return
	}

	// Проверяем существование и доступность директории
	err = ValidateDirectory(targetDir)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		if flags.NoError {
			os.Exit(1)
		}
		return
	}

	// Изменяем текущую директорию
	err = os.Chdir(targetDir)
	if err != nil {
		fmt.Printf("Ошибка: не удаётся изменить директорию на '%s': %v\n", targetDir, err)
		os.Exit(1)
	}

	// Запускаем shell с обновленным OLDPWD
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.Command(shell)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Устанавливаем OLDPWD для shell
	env := os.Environ()
	env = append(env, fmt.Sprintf("OLDPWD=%s", currentDir))
	cmd.Env = env

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
