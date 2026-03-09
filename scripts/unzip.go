package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UnzipFlags содержит флаги управления unzip
type UnzipFlags struct {
	Quiet bool   // -q тихий режим
	List  bool   // -l показать содержимое
	Dir   string // -d директория для извлечения
	Help  bool   // -h справка
}

// showHelp выводит справку по использованию
func showHelp() {
	fmt.Println(`Использование: ./unzip [КЛЮЧИ] arch.zip

Ключи:
  -q         тихий режим
  -l         показать содержимое архива
  -d DIR     извлекать в директорию DIR
  -h         показать эту справку
  --help     показать эту справку

Описание:
  unzip - распаковывает ZIP архивы.

Примеры:
  ./unzip arch.zip                 Распаковать в текущую директорию
  ./unzip -d dir/ arch.zip         Распаковать в директорию dir/
  ./unzip -l arch.zip              Показать содержимое
  ./unzip -lq arch.zip             Показать содержимое без деталей`)
}

// parseFlags парсит флаги из аргументов командной строки
func parseFlags(args []string) (UnzipFlags, string, error) {
	flags := UnzipFlags{}
	var archiveName string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") && arg != "-" {
			if arg == "--help" || arg == "-h" {
				flags.Help = true
			} else if len(arg) > 2 {
				for _, ch := range arg[1:] {
					switch ch {
					case 'q':
						flags.Quiet = true
					case 'l':
						flags.List = true
					case 'd':
						if i+1 < len(args) {
							flags.Dir = args[i+1]
							i++
						}
					case 'h':
						flags.Help = true
					default:
						return flags, "", fmt.Errorf("неизвестный флаг: -%c", ch)
					}
				}
			} else {
				switch arg {
				case "-q":
					flags.Quiet = true
				case "-l":
					flags.List = true
				case "-d":
					if i+1 < len(args) {
						flags.Dir = args[i+1]
						i++
					}
				case "-h":
					flags.Help = true
				default:
					return flags, "", fmt.Errorf("неизвестный флаг: %s", arg)
				}
			}
		} else {
			archiveName = arg
		}
	}

	return flags, archiveName, nil
}

// listArchive показывает содержимое архива
func listArchive(archiveName string, quiet bool) error {
	r, err := zip.OpenReader(archiveName)
	if err != nil {
		return fmt.Errorf("cannot find or open %s", archiveName)
	}
	defer r.Close()

	if !quiet {
		fmt.Printf("Archive:  %s\n", archiveName)
		fmt.Printf("  Length      Date    Time    Name\n")
		fmt.Printf("---------  ---------- -----   ----\n")
	}

	totalSize := int64(0)
	fileCount := 0

	for _, f := range r.File {
		size := f.FileInfo().Size()
		totalSize += size
		fileCount++

		if !quiet {
			fmt.Printf("     %6d  %11s %5s   %s\n",
				size,
				f.Modified.Format("01-02-2006"),
				f.Modified.Format("15:04"),
				f.Name)
		}
	}

	if !quiet {
		fmt.Printf("---------                     -------\n")
		fmt.Printf("     %6d                     %d files\n", totalSize, fileCount)
	}

	return nil
}

// extractArchive распаковывает архив
func extractArchive(archiveName string, quiet bool, targetDir string) error {
	r, err := zip.OpenReader(archiveName)
	if err != nil {
		return fmt.Errorf("cannot find or open %s", archiveName)
	}
	defer r.Close()

	// Создаём целевую директорию если нужно
	if targetDir != "" {
		os.MkdirAll(targetDir, 0755)
	}

	extracted := 0

	for _, f := range r.File {
		filePath := f.Name
		if targetDir != "" {
			filePath = filepath.Join(targetDir, f.Name)
		}

		// Директории
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, f.Mode())
			continue
		}

		// Создаём родительские директории
		os.MkdirAll(filepath.Dir(filePath), 0755)

		// Открываем файл для записи
		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			if !quiet {
				fmt.Printf("unzip: cannot create %s: %v\n", filePath, err)
			}
			continue
		}

		// Копируем содержимое
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			continue
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err == nil {
			extracted++
			if !quiet {
				fmt.Printf("  extracting: %s\n", f.Name)
			}
		}
	}

	if !quiet {
		fmt.Printf("Archive:  %s\n", archiveName)
		fmt.Printf("%8d files\n", extracted)
	}

	return nil
}

func main() {
	flags, archiveName, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}

	if flags.Help {
		showHelp()
		return
	}

	if archiveName == "" {
		fmt.Fprintln(os.Stderr, "Ошибка: необходимо указать архив")
		fmt.Fprintln(os.Stderr, "Используй ./unzip -h для справки")
		os.Exit(1)
	}

	var e error

	if flags.List {
		e = listArchive(archiveName, flags.Quiet)
	} else {
		e = extractArchive(archiveName, flags.Quiet, flags.Dir)
	}

	if e != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", e)
		os.Exit(1)
	}
}
