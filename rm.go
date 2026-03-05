package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RmFlags содержит флаги для управления удалением файлов и директорий
type RmFlags struct {
	Recursive bool // -R рекурсивное удаление директорий
	Force     bool // -f принудительное удаление без ошибок
	Verbose   bool // -v подробный вывод о каждом удалённом файле
	Help      bool // -h/--help справка
}

// ShowHelp выводит справку по использованию команды rm
func ShowHelp() {
	help := `
Использование: ./rm [КЛЮЧИ] путь [путь2] [путь3] ...

Ключи:
  -R         Рекурсивное удаление директорий и их содержимого
  -f         Принудительное удаление (игнорировать несуществующие файлы)
  -v         Подробный вывод о каждом удалённом файле
  -h         Показать эту справку
  --help     Показать эту справку

Описание:
  rm - удаляет файлы и директории.
  
  Команда rm используется для удаления файлов и директорий.
  Без флага -R удаляет только файлы.
  С флагом -R может удалять директории рекурсивно.
  С флагом -f игнорирует ошибки и не выводит сообщения о несуществующих файлах.

Примеры:
  ./rm file.txt                  Удалить файл
  ./rm -v file.txt               Удалить файл с выводом
  ./rm -R dir/                   Удалить директорию рекурсивно
  ./rm -Rf dir/                  Удалить директорию принудительно
  ./rm -Rfv dir/                 Удалить с выводом и без ошибок
  ./rm file1.txt file2.txt       Удалить несколько файлов
  ./rm -f nonexistent.txt        Удалить без ошибки если не существует
`
	fmt.Println(help)
}

// ParseRmFlags парсит флаги для rm
func ParseRmFlags(args []string) (RmFlags, []string, error) {
	flags := RmFlags{}
	var paths []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			flags.Help = true
		} else if strings.HasPrefix(arg, "-") && arg != "-" {
			// Обработка комбинированных флагов
			for j := 1; j < len(arg); j++ {
				ch := rune(arg[j])
				switch ch {
				case 'R':
					flags.Recursive = true
				case 'f':
					flags.Force = true
				case 'v':
					flags.Verbose = true
				case 'h':
					flags.Help = true
				default:
					return flags, nil, fmt.Errorf("неизвестный флаг: -%c", ch)
				}
			}
		} else {
			paths = append(paths, arg)
		}
	}

	return flags, paths, nil
}

// ValidatePath проверяет существование пути
func ValidatePath(path string, force bool) error {
	_, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if force {
				return nil
			}
			return fmt.Errorf("не удалось удалить '%s': Нет такого файла или каталога", path)
		}
		return fmt.Errorf("ошибка доступа к '%s': %v", path, err)
	}

	return nil
}

// RemoveFile удаляет один файл
func RemoveFile(path string, verbose bool) error {
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("не удалось удалить '%s': %v", path, err)
	}

	if verbose {
		fmt.Printf("удалён '%s'\n", path)
	}

	return nil
}

// RemoveDirectoryRecursive удаляет директорию рекурсивно
func RemoveDirectoryRecursive(path string, verbose bool) error {
	if verbose {
		return RemoveDirectoryRecursiveVerbose(path)
	}

	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("не удалось удалить '%s': %v", path, err)
	}

	return nil
}

// RemoveDirectoryRecursiveVerbose удаляет директорию рекурсивно с выводом
func RemoveDirectoryRecursiveVerbose(path string) error {
	var filesToRemove []string

	// Собираем все файлы и директории
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		filesToRemove = append(filesToRemove, p)
		return nil
	})

	if err != nil {
		return fmt.Errorf("не удалось обойти '%s': %v", path, err)
	}

	// Удаляем в обратном порядке
	for i := len(filesToRemove) - 1; i >= 0; i-- {
		p := filesToRemove[i]
		err := os.Remove(p)
		if err != nil {
			return fmt.Errorf("не удалось удалить '%s': %v", p, err)
		}
		fmt.Printf("удалён '%s'\n", p)
	}

	return nil
}

// RemovePath удаляет файл или директорию
func RemovePath(path string, flags RmFlags) error {
	// Проверяем существование
	err := ValidatePath(path, flags.Force)
	if err != nil {
		return err
	}

	// Получаем информацию о пути
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) && flags.Force {
			return nil
		}
		return fmt.Errorf("не удалось получить информацию о '%s': %v", path, err)
	}

	// Если это директория
	if info.IsDir() {
		if !flags.Recursive {
			return fmt.Errorf("не удалось удалить '%s': Это каталог", path)
		}
		return RemoveDirectoryRecursive(path, flags.Verbose)
	}

	// Удаляем файл
	return RemoveFile(path, flags.Verbose)
}

func main() {
	flags, paths, err := ParseRmFlags(os.Args[1:])
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("Используй ./rm --help для справки")
		os.Exit(1)
	}

	if flags.Help {
		ShowHelp()
		return
	}

	if len(paths) == 0 {
		if !flags.Force {
			fmt.Println("rm: отсутствует операнд")
			fmt.Println("Используй ./rm --help для справки")
			os.Exit(1)
		}
		return
	}

	// Удаляем все пути
	exitCode := 0
	for _, path := range paths {
		err := RemovePath(path, flags)

		if err != nil {
			if !flags.Force {
				fmt.Printf("rm: %v\n", err)
				exitCode = 1
			}
		}
	}

	os.Exit(exitCode)
}
