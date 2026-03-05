package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DuOptions содержит флаги для управления du
type DuOptions struct {
	Help      bool
	HumanRead bool // -h в удобном виде
	Total     bool // -c выводить общий итог
	All       bool // -a для всех файлов, не только каталогов
	Bytes     bool // -b размер в байтах
	Paths     []string
}

// DuResult содержит результат для одного пути
type DuResult struct {
	Size int64  // Размер в килобайтах
	Path string // Путь к файлу/каталогу
}

// showHelp выводит справку по использованию команды
func showHelp() {
	fmt.Println(`Использование: du [КЛЮЧИ] [ФАЙЛ/КАТАЛОГ]...

Ключи:
  -h           печатать размеры в удобном для человека виде (K, M, G)
  -c           выводить общий итог в конце
  -a           печатать объём для всех файлов, а не только каталогов
  -b           эквивалентно '--apparent-size --block-size=1'
  --help       показать эту справку

Примеры:
  ./du dir/                  Размер каталога и подкаталогов
  ./du -h dir/               Размер в удобном виде
  ./du -c dir1/ dir2/        Размер и общий итог
  ./du -a dir/               Размер всех файлов и каталогов
  ./du -b dir/               Размер в байтах`)
}

// parseFlags парсит аргументы командной строки
func parseFlags(args []string) (DuOptions, error) {
	opts := DuOptions{}
	paths := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Проверяем, что это флаг
		if strings.HasPrefix(arg, "-") && arg != "-" {
			if arg == "--help" {
				opts.Help = true
			} else if strings.HasPrefix(arg, "--") {
				return opts, fmt.Errorf("неизвестный флаг: %s", arg)
			} else {
				for _, ch := range arg[1:] {
					switch ch {
					case 'h':
						opts.HumanRead = true
					case 'c':
						opts.Total = true
					case 'a':
						opts.All = true
					case 'b':
						opts.Bytes = true
					default:
						return opts, fmt.Errorf("неизвестный флаг: -%c", ch)
					}
				}
			}
		} else {
			paths = append(paths, arg)
		}
	}

	opts.Paths = paths
	return opts, nil
}

// sizeInKB преобразует размер в байтах в килобайты
func sizeInKB(bytes int64) int64 {
	if bytes == 0 {
		return 1
	}
	return (bytes + 512) / 1024
}

// formatSize форматирует размер в удобный вид
func formatSize(sizeKB int64, humanRead bool, inBytes bool) string {
	if inBytes {
		return fmt.Sprintf("%d", sizeKB*1024)
	}

	if !humanRead {
		return fmt.Sprintf("%d", sizeKB)
	}

	// Для человеческого формата конвертируем KB в байты для форматирования
	return formatHumanReadable1024(sizeKB * 1024)
}

// formatHumanReadable1024 форматирует размер в понятном формате степени 1024
func formatHumanReadable1024(sizeBytes int64) string {
	units := []string{"B", "K", "M", "G", "T"}
	divisor := 1024.0
	sizeFloat := float64(sizeBytes)

	for i := 0; i < len(units); i++ {
		if sizeFloat < divisor || i == len(units)-1 {
			if i == 0 {
				return fmt.Sprintf("%.0f%s", sizeFloat, units[i])
			}
			if sizeFloat == float64(int64(sizeFloat)) {
				return fmt.Sprintf("%.0f%s", sizeFloat, units[i])
			}
			return fmt.Sprintf("%.1f%s", sizeFloat, units[i])
		}
		sizeFloat /= divisor
	}

	return fmt.Sprintf("%.0f", sizeFloat)
}

// getDirSize рекурсивно подсчитывает размер каталога в килобайтах
func getDirSize(path string, opts DuOptions, results *[]DuResult) (int64, error) {
	var totalSize int64

	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}

	// Добавляем размер самого каталога
	totalSize = 4

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if entry.IsDir() {
			// Рекурсивно подсчитываем размер подкаталога
			subSize, err := getDirSize(fullPath, opts, results)
			if err != nil {
				continue
			}
			totalSize += subSize

			// Всегда выводим размер подкаталога
			*results = append(*results, DuResult{
				Size: subSize,
				Path: fullPath,
			})
		} else {
			// Это файл добавляем его размер с округлением до 4
			fileBytes := info.Size()
			// Округляем размер файла до блоков 4
			fileBlocks := (fileBytes + 4095) / 4096
			fileSize := fileBlocks * 4
			totalSize += fileSize

			// Выводим размер файла только если указан флаг -a
			if opts.All {
				*results = append(*results, DuResult{
					Size: fileSize,
					Path: fullPath,
				})
			}
		}
	}

	return totalSize, nil
}

// processPath обрабатывает один путь
func processPath(path string, opts DuOptions) ([]DuResult, int64, error) {
	var results []DuResult
	var topLevelSize int64

	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, err
	}

	if info.IsDir() {
		// Это каталог
		size, err := getDirSize(path, opts, &results)
		if err != nil {
			return nil, 0, err
		}

		// Добавляем сам корневой каталог в конец
		results = append(results, DuResult{
			Size: size,
			Path: path,
		})

		topLevelSize = size
	} else {
		size := sizeInKB(info.Size())
		results = append(results, DuResult{
			Size: size,
			Path: path,
		})

		topLevelSize = size
	}

	return results, topLevelSize, nil
}

// printResults выводит результаты
func printResults(results []DuResult, topLevelPaths map[string]int64, opts DuOptions) {
	for _, result := range results {
		size := result.Size

		sizeStr := formatSize(size, opts.HumanRead, opts.Bytes)
		fmt.Printf("%s\t%s\n", sizeStr, result.Path)
	}

	// Выводим общий итог если указан флаг -c
	if opts.Total && len(topLevelPaths) > 0 {
		var totalSize int64
		for _, size := range topLevelPaths {
			totalSize += size
		}
		sizeStr := formatSize(totalSize, opts.HumanRead, opts.Bytes)
		fmt.Printf("%s\ttotal\n", sizeStr)
	}
}

func main() {
	opts, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "du: %v\n", err)
		os.Exit(1)
	}

	if opts.Help {
		showHelp()
		return
	}

	// Если пути не указаны используем текущий каталог
	if len(opts.Paths) == 0 {
		opts.Paths = []string{"."}
	}

	var allResults []DuResult
	topLevelPaths := make(map[string]int64)

	for _, path := range opts.Paths {
		results, topLevelSize, err := processPath(path, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "du: не могу получить доступ к '%s': %v\n", path, err)
			continue
		}
		allResults = append(allResults, results...)
		topLevelPaths[path] = topLevelSize
	}

	printResults(allResults, topLevelPaths, opts)
}
