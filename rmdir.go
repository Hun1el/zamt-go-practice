package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RmdirFlags содержит флаги для управления удалением директорий
type RmdirFlags struct {
	Parents              bool // -p удалять родительские директории если они пустые
	Verbose              bool // -v подробный вывод о каждой удалённой директории
	IgnoreFailOnNonEmpty bool // --ignore-fail-on-non-empty игнорировать ошибки непустых директорий
	Version              bool // --version показать версию программы
	Help                 bool // -h/--help справка
}

// ShowHelp выводит справку по использованию команды rmdir
func ShowHelp() {
	help := `
Использование: rmdir [КЛЮЧИ] путь [путь2] [путь3] ...

Ключи:
  -p                        	Удалять родительские директории если они пустые
  -v                        	Подробный вывод о каждой удалённой директории
  --ignore-fail-on-non-empty 	Игнорировать ошибки при удалении непустых директорий
  --version                 	Показать версию программы
  -h                        	Показать эту справку
  --help                    	Показать эту справку

Описание:
  rmdir - удаляет одну или несколько пустых директорий.

Примеры:
  ./rmdir empty_folder                          Удалить пустую директорию
  ./rmdir -p a/b/c                              Удалить c и его пустые родители
  ./rmdir -v empty_folder                       Удалить с подробным выводом
  ./rmdir -pv a/b/c                             Удалить родителей с подробным выводом
  ./rmdir folder1 folder2                       Удалить несколько директорий
  ./rmdir --ignore-fail-on-non-empty dir1 dir2  Игнорировать ошибки непустых директорий
`
	fmt.Println(help)
}

// GetVersion возвращает версию программы
func GetVersion() string {
	return "rmdir (Go implementation) version 1.0.0"
}

// ParseRmdirFlags парсит флаги для rmdir
func ParseRmdirFlags(args []string) (RmdirFlags, []string, error) {
	flags := RmdirFlags{}
	var paths []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			flags.Help = true
		} else if arg == "--version" {
			flags.Version = true
		} else if arg == "--ignore-fail-on-non-empty" {
			// Флаг для игнорирования ошибок непустых директорий
			flags.IgnoreFailOnNonEmpty = true
		} else if strings.HasPrefix(arg, "-") && arg != "-" {
			// Обработка комбинированных флагов
			for j := 1; j < len(arg); j++ {
				ch := rune(arg[j])
				switch ch {
				case 'p':
					flags.Parents = true
				case 'v':
					flags.Verbose = true
				case 'h':
					flags.Help = true
				default:
					return flags, nil, fmt.Errorf("неизвестный флаг: -%c", ch)
				}
			}
		} else {
			// Это путь
			paths = append(paths, arg)
		}
	}

	return flags, paths, nil
}

// IsDirectoryEmpty проверяет пуста ли директория
func IsDirectoryEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}

	return len(entries) == 0, nil
}

// ValidateDirectory проверяет существование и является ли путь директорией
func ValidateDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("ошибка: не удаётся получить доступ к '%s': %v", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("ошибка: '%s' не является директорией", path)
	}

	return nil
}

// RemoveDirectory удаляет директорию
func RemoveDirectory(path string, verbose bool, ignoreFailOnNonEmpty bool) error {
	// Проверяем существование и что это директория
	err := ValidateDirectory(path)
	if err != nil {
		return err
	}

	// Проверяем пуста ли директория
	isEmpty, err := IsDirectoryEmpty(path)
	if err != nil {
		return fmt.Errorf("ошибка при проверке директории '%s': %v", path, err)
	}

	if !isEmpty {
		// Если флаг --ignore-fail-on-non-empty установлен, игнорируем ошибку
		if ignoreFailOnNonEmpty {
			if verbose {
				fmt.Printf("Директория '%s' не пуста - пропускаем\n", path)
			}
			return nil
		}
		return fmt.Errorf("ошибка: директория '%s' не пуста (rmdir удаляет только пустые директории)", path)
	}

	// Удаляем директорию
	err = os.Remove(path)
	if err != nil {
		return fmt.Errorf("ошибка при удалении директории '%s': %v", path, err)
	}

	if verbose {
		fmt.Printf("Директория '%s' успешно удалена\n", path)
	}

	return nil
}

// RemoveDirectoryWithParents удаляет директорию и её пустые родители
func RemoveDirectoryWithParents(path string, verbose bool, ignoreFailOnNonEmpty bool) error {
	path = filepath.Clean(path)

	err := RemoveDirectory(path, verbose, ignoreFailOnNonEmpty)
	if err != nil {
		return err
	}

	// Удаляем родительские директории если они пустые
	parent := filepath.Dir(path)
	for parent != path && parent != "/" && parent != "." {
		// Проверяем что это директория
		info, err := os.Stat(parent)
		if err != nil {
			break
		}

		if !info.IsDir() {
			break
		}

		// Проверяем пуста ли родительская директория
		isEmpty, err := IsDirectoryEmpty(parent)
		if err != nil {
			break
		}

		if !isEmpty {
			break
		}

		err = os.Remove(parent)
		if err != nil {
			break
		}

		if verbose {
			fmt.Printf("Директория '%s' успешно удалена\n", parent)
		}

		parent = filepath.Dir(parent)
	}

	return nil
}

func main() {
	flags, paths, err := ParseRmdirFlags(os.Args[1:])
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("Используй ./rmdir --help для справки")
		os.Exit(1)
	}

	if flags.Version {
		fmt.Println(GetVersion())
		os.Exit(0)
	}

	if flags.Help {
		ShowHelp()
		return
	}

	if len(paths) == 0 {
		fmt.Printf("Ошибка: необходимо указать хотя бы один путь\n")
		fmt.Println("Используй ./rmdir --help для справки")
		os.Exit(1)
	}

	// Удаляем все директории
	exitCode := 0
	for _, path := range paths {
		var err error

		if flags.Parents {
			err = RemoveDirectoryWithParents(path, flags.Verbose, flags.IgnoreFailOnNonEmpty)
		} else {
			err = RemoveDirectory(path, flags.Verbose, flags.IgnoreFailOnNonEmpty)
		}

		if err != nil {
			fmt.Printf("%v\n", err)
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}
