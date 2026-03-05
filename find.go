package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FindFlags содержит флаги для команды find
type FindFlags struct {
	Path        string // путь для поиска
	NamePattern string // -name ПАТТЕРН
	FileType    string // -type ТИП (f, d, l)
	MaxDepth    int    // -maxdepth N
	Help        bool   // -h/--help справка
}

// ShowHelp выводит справку
func ShowHelp() {
	help := `
Использование: find [ПУТЬ] [КЛЮЧИ]


Ключи:
  -name ПАТТЕРН     искать по имени файла (glob паттерн)
  -type ТИП         f (обычный файл), d (директория), l (символическая ссылка)
  -maxdepth N       максимальная глубина поиска (N >= 0)
  -h, --help        показать эту справку


Описание:
  find - рекурсивно ищет файлы в директориях.

Примеры:
  find                           все файлы от текущей директории
  find /tmp                      все файлы в /tmp
  find . -name "*.txt"           все .txt файлы в текущем месте
  find / -name "*.txt"           все .txt файлы везде
  find /home -name "*.txt"       все .txt файлы в каталоге /home
  find . -type f                 только обычные файлы
  find . -type d                 только директории
  find . -maxdepth 0             только сам путь (.)
  find . -maxdepth 2             глубина максимум 2 уровня
  find . -name "*.go" -type f    .go файлы
`
	fmt.Println(help)
}

// ParseFindFlags парсит флаги для команды find
func ParseFindFlags(args []string) (FindFlags, error) {
	flags := FindFlags{
		Path:     ".",
		MaxDepth: -1,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-h" || arg == "--help" {
			flags.Help = true
		} else if arg == "-name" {
			if i+1 >= len(args) {
				return flags, fmt.Errorf("флаг -name требует аргумента")
			}
			flags.NamePattern = args[i+1]
			i++
		} else if arg == "-type" {
			if i+1 >= len(args) {
				return flags, fmt.Errorf("флаг -type требует аргумента")
			}
			flags.FileType = args[i+1]
			i++
		} else if arg == "-maxdepth" {
			if i+1 >= len(args) {
				return flags, fmt.Errorf("флаг -maxdepth требует аргумента")
			}
			depth, err := strconv.Atoi(args[i+1])
			if err != nil || depth < 0 {
				return flags, fmt.Errorf("maxdepth должен быть неотрицательным числом")
			}
			flags.MaxDepth = depth
			i++
		} else if strings.HasPrefix(arg, "-") {
			// Неизвестный флаг
			return flags, fmt.Errorf("неизвестный флаг: %s", arg)
		} else {
			// Аргумент без дефиса это путь
			// существует ли он и является ли директорией
			info, err := os.Stat(arg)
			if err != nil {
				return flags, fmt.Errorf("paths must precede expression: `%s'", arg)
			}
			if !info.IsDir() {
				return flags, fmt.Errorf("paths must precede expression: `%s'\nfind: possible unquoted pattern after predicate `-name'?", arg)
			}
			flags.Path = arg
		}
	}

	return flags, nil
}

// MatchesPattern проверяет совпадение с паттерном
func MatchesPattern(pattern, name string) bool {
	matched, _ := filepath.Match(pattern, name)
	return matched
}

// MatchesType проверяет тип файла
func MatchesType(typeFilter string, info fs.FileInfo) bool {
	switch typeFilter {
	case "f":
		return info.Mode().IsRegular()
	case "d":
		return info.IsDir()
	case "l":
		return (info.Mode() & fs.ModeSymlink) != 0
	default:
		return true
	}
}

// shouldPrintPath проверяет должен ли путь быть выведен
func shouldPrintPath(name string, fileType string, info fs.FileInfo, namePattern string) bool {
	// Проверяем -type
	if fileType != "" && !MatchesType(fileType, info) {
		return false
	}

	// Проверяем -name
	if namePattern != "" && !MatchesPattern(namePattern, name) {
		return false
	}

	return true
}

// printPath выводит путь в правильном формате
func printPath(path string, startPath string) {
	if startPath == "." || startPath == "./" {
		// Выводим с префиксом ./
		if path == "." {
			fmt.Println(".")
		} else {
			normalizedPath := filepath.ToSlash(path)
			if !strings.HasPrefix(normalizedPath, "./") {
				normalizedPath = "./" + normalizedPath
			}
			fmt.Println(normalizedPath)
		}
	} else {
		fmt.Println(path)
	}
}

// Find рекурсивно ищет файлы
func Find(path string, flags FindFlags) error {
	// Абсолютный путь для расчёта глубины
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	// ВЫводим сам корневой путь
	rootInfo, err := os.Stat(path)
	if err == nil {
		if shouldPrintPath(filepath.Base(path), flags.FileType, rootInfo, flags.NamePattern) {
			printPath(path, path)
		}

		// Если maxdepth 0 выводим только сам путь и выходим
		if flags.MaxDepth == 0 {
			return nil
		}
	}

	// Обходим дерево
	err = filepath.WalkDir(path, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			fmt.Fprintf(os.Stderr, "ошибка при чтении %s: %v\n", currentPath, walkErr)
			return nil
		}

		// Пропускаем сам корневой путь
		if currentPath == path {
			return nil
		}

		// Получаем информацию о файле
		info, err := d.Info()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ошибка при получении информации о %s: %v\n", currentPath, err)
			return nil
		}

		// Вычисляем глубину
		relPath, _ := filepath.Rel(absPath, currentPath)
		var depth int
		if relPath == "." {
			depth = 0
		} else {
			parts := strings.Split(relPath, string(filepath.Separator))
			depth = len(parts)
		}

		// Проверяем maxdepth
		if flags.MaxDepth >= 0 && depth > flags.MaxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Проверяем фильтры и выводим если подходит
		if shouldPrintPath(info.Name(), flags.FileType, info, flags.NamePattern) {
			printPath(currentPath, path)
		}

		// Если maxdepth достигнут и это директория, пропускаем её содержимое
		if flags.MaxDepth >= 0 && depth >= flags.MaxDepth && d.IsDir() {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("ошибка при обходе директории: %v", err)
	}

	return nil
}

func main() {
	flags, err := ParseFindFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "find: %v\n", err)
		os.Exit(1)
	}

	if flags.Help {
		ShowHelp()
		return
	}

	// Выполняем поиск
	if err := Find(flags.Path, flags); err != nil {
		fmt.Fprintf(os.Stderr, "find: %v\n", err)
		os.Exit(1)
	}
}
