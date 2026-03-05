package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

// FileInfo содержит информацию о файле для вывода
type FileInfo struct {
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
	Mode    os.FileMode
	Path    string
	Nlink   uint64
	Blocks  int64
}

// Flags содержит флаги для управления выводом
type Flags struct {
	Recursive bool
	All       bool
	Long      bool
	Reverse   bool
	Human     bool
	Help      bool
}

// FormatSize преобразует размер в человеческий формат
func FormatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	}

	divisor := float64(1024)
	units := []string{"K", "M", "G", "T"}

	for _, unit := range units {
		if float64(size) < divisor*1024 {
			return fmt.Sprintf("%.1f%s", float64(size)/divisor, unit)
		}

		divisor *= 1024
	}

	return fmt.Sprintf("%.1fP", float64(size)/divisor)
}

// FormatPermissions форматирует права доступа в виде строки
func FormatPermissions(mode os.FileMode) string {
	permissions := ""

	// Владелец
	if mode&0400 != 0 {
		permissions += "r"
	} else {
		permissions += "-"
	}
	if mode&0200 != 0 {
		permissions += "w"
	} else {
		permissions += "-"
	}
	if mode&0100 != 0 {
		permissions += "x"
	} else {
		permissions += "-"
	}

	// Группа
	if mode&0040 != 0 {
		permissions += "r"
	} else {
		permissions += "-"
	}
	if mode&0020 != 0 {
		permissions += "w"
	} else {
		permissions += "-"
	}
	if mode&0010 != 0 {
		permissions += "x"
	} else {
		permissions += "-"
	}

	// Остальные
	if mode&0004 != 0 {
		permissions += "r"
	} else {
		permissions += "-"
	}
	if mode&0002 != 0 {
		permissions += "w"
	} else {
		permissions += "-"
	}
	if mode&0001 != 0 {
		permissions += "x"
	} else {
		permissions += "-"
	}

	return permissions
}

// ShowHelp выводит справку по использованию
func ShowHelp() {
	help := `
Использование: ./ls [КЛЮЧИ] [путь]

Ключи:
  -a         Показывать все файлы, включая скрытые
  -l         Подробный формат вывода (права, размер, дата, имя)
  -R         Рекурсивный вывод содержимого всех поддиректорий
  -r         Обратный порядок сортировки (Z-A вместо A-Z)
  -h         Размер в человеческом формате (1.0K, 3.0M и т.д.)
  --help     Показать эту справку

Описание:
  ls - выводит содержимое директории в различных форматах.
  
  Команда ls используется для просмотра файлов и папок в директории.

Примеры:
  ./ls                      Вывести содержимое текущей директории
  ./ls -l                   Подробный список текущей директории
  ./ls -l -a                Подробный список со скрытыми файлами
  ./ls -R -h                Рекурсивный вывод с размерами в формате 1.0K
  ./ls -l -r /home          Список /home в обратном порядке
`
	fmt.Println(help)
}

// ParseFlags парсит командную строку и выделяет флаги и путь
func ParseFlags(args []string) (Flags, string, error) {
	flags := Flags{}
	path := "."

	for _, arg := range args {
		if arg == "--help" {
			flags.Help = true
		} else if strings.HasPrefix(arg, "-") && arg != "-" {
			for _, ch := range arg[1:] {
				switch ch {
				case 'R':
					flags.Recursive = true
				case 'a':
					flags.All = true
				case 'l':
					flags.Long = true
				case 'r':
					flags.Reverse = true
				case 'h':
					flags.Human = true
				default:
					return flags, "", fmt.Errorf("неизвестный флаг: -%c", ch)
				}
			}
		} else {
			path = arg
		}
	}

	return flags, path, nil
}

// IsHidden проверяет, является ли файл скрытым
func IsHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

// getStatInfo получает Nlink и Blocks из syscall.Stat_t
func getStatInfo(info os.FileInfo) (uint64, int64) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if ok {
		return uint64(stat.Nlink), stat.Blocks
	}
	return 1, 0
}

// ReadDir читает содержимое директории
func ReadDir(path string, showHidden bool) ([]FileInfo, error) {
	entries, err := os.ReadDir(path)

	if err != nil {
		return nil, err
	}

	var files []FileInfo

	if showHidden {
		// Добавляем текущую директорию
		currentDir, err := os.Stat(path)
		if err == nil {
			nlink, blocks := getStatInfo(currentDir)
			files = append(files, FileInfo{
				Name:    ".",
				IsDir:   true,
				Size:    currentDir.Size(),
				ModTime: currentDir.ModTime(),
				Mode:    currentDir.Mode(),
				Path:    path,
				Nlink:   nlink,
				Blocks:  blocks,
			})
		}

		absPath, err := filepath.Abs(path)
		if err == nil {
			parentPath := filepath.Dir(absPath)
			parentDir, err := os.Stat(parentPath)
			if err == nil {
				nlink, blocks := getStatInfo(parentDir)
				files = append(files, FileInfo{
					Name:    "..",
					IsDir:   true,
					Size:    parentDir.Size(),
					ModTime: parentDir.ModTime(),
					Mode:    parentDir.Mode(),
					Path:    parentPath,
					Nlink:   nlink,
					Blocks:  blocks,
				})
			}
		}
	}

	// Используем срез для хранения информации о файлах
	for _, entry := range entries {
		name := entry.Name()

		if !showHidden && IsHidden(name) {
			continue
		}

		info, err := entry.Info()

		if err != nil {
			continue
		}

		nlink, blocks := getStatInfo(info)

		files = append(files, FileInfo{
			Name:    name,
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Mode:    info.Mode(),
			Path:    filepath.Join(path, name),
			Nlink:   nlink,
			Blocks:  blocks,
		})
	}

	return files, nil
}

// PrintLongFormat выводит информацию о файле в подробном формате
func PrintLongFormat(file FileInfo, useHuman bool) {
	// Определяем символ для типа файла
	typeChar := "-"

	if file.IsDir {
		typeChar = "d"
	}

	// Форматируем размер
	sizeStr := fmt.Sprintf("%d", file.Size)

	if useHuman {
		sizeStr = FormatSize(file.Size)
	}

	// Форматируем дату и время в русском формате
	months := map[string]string{
		"Jan": "янв", "Feb": "фев", "Mar": "мар", "Apr": "апр",
		"May": "май", "Jun": "июн", "Jul": "июл", "Aug": "авг",
		"Sep": "сен", "Oct": "окт", "Nov": "ноя", "Dec": "дек",
	}

	enFormat := file.ModTime.Format("Jan 02 15:04")
	parts := strings.Fields(enFormat)
	if len(parts) >= 3 {
		ruMonth := months[parts[0]]
		timeStr := fmt.Sprintf("%s %s %s", ruMonth, parts[1], parts[2])
		perms := FormatPermissions(file.Mode)

		// Выводим информацию с количеством ссылок
		fmt.Printf("%s%s. %d user user %10s %s %s\n", typeChar, perms, file.Nlink, sizeStr, timeStr, file.Name)
	}
}

// SortFiles сортирует массив файлов по имени
func SortFiles(files []FileInfo, reverse bool) {
	sort.Slice(files, func(i, j int) bool {
		if reverse {
			return files[i].Name > files[j].Name
		}
		return files[i].Name < files[j].Name
	})
}

// CalculateTotalBlocks вычисляет общее количество блоков
func CalculateTotalBlocks(files []FileInfo, useHuman bool) string {
	var totalBlocks int64

	for _, file := range files {
		totalBlocks += file.Blocks
	}

	totalBlocks = totalBlocks / 2

	// Если флаг -h установлен
	if useHuman {
		return FormatSize(totalBlocks * 1024)
	}

	return fmt.Sprintf("%d", totalBlocks)
}

// ListDirectory выводит содержимое директории
func ListDirectory(path string, flags Flags) error {
	// Читаем содержимое директории
	files, err := ReadDir(path, flags.All)

	if err != nil {
		return fmt.Errorf("ошибка при чтении директории '%s': %v", path, err)
	}

	// Сортируем файлы
	SortFiles(files, flags.Reverse)

	// Выводим содержимое
	if flags.Long {
		// Вычисляем общее количество блоков
		totalBlocksStr := CalculateTotalBlocks(files, flags.Human)

		// Выводим итого в начале
		fmt.Printf("итого %s\n", totalBlocksStr)

		// Подробный формат
		for _, file := range files {
			PrintLongFormat(file, flags.Human)
		}
	} else {
		// Если -R установлен выводим название директории
		if flags.Recursive {
			fmt.Printf(".:\n")
		}

		var names []string
		for _, file := range files {
			if file.Name == "." || file.Name == ".." {
				continue
			}

			if file.IsDir {
				names = append(names, file.Name+"/")
			} else {
				names = append(names, file.Name)
			}
		}
		fmt.Println(strings.Join(names, " "))
	}

	// Если флаг -R установлен рекурсивно выводим содержимое поддиректорий
	if flags.Recursive {
		visited := make(map[string]bool)
		RecursiveList(path, flags, visited)
	}

	return nil
}

// RecursiveList рекурсивно выводит содержимое всех поддиректорий
func RecursiveList(path string, flags Flags, visited map[string]bool) error {
	absPath, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	if visited[absPath] {
		return nil
	}

	visited[absPath] = true

	files, err := ReadDir(path, flags.All)

	if err != nil {
		return err
	}

	SortFiles(files, flags.Reverse)

	for _, file := range files {
		if file.IsDir {
			fmt.Printf("\n%s:\n", file.Path)

			// Рекурсивно выводим содержимое поддиректории
			subFiles, err := ReadDir(file.Path, flags.All)
			if err != nil {
				continue
			}

			SortFiles(subFiles, flags.Reverse)

			// Выводим в одну строку
			var names []string
			for _, subFile := range subFiles {
				if subFile.IsDir {
					names = append(names, subFile.Name+"/")
				} else {
					names = append(names, subFile.Name)
				}
			}

			if len(names) > 0 {
				fmt.Println(strings.Join(names, " "))
			}

			RecursiveList(file.Path, flags, visited)
		}
	}

	return nil
}

func main() {
	// Парсим аргументы командной строки
	flags, path, err := ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("Используй ./ls --help для справки")
		os.Exit(1)
	}

	if flags.Help {
		ShowHelp()
		return
	}

	// Проверяем корректность пути
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Ошибка: не удаётся получить доступ к '%s': %v\n", path, err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Printf("Ошибка: '%s' не является директорией\n", path)
		os.Exit(1)
	}

	// Выводим содержимое директории
	err = ListDirectory(path, flags)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		os.Exit(1)
	}
}
