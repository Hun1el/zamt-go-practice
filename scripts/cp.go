package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CpFlags struct {
	Help        bool // -h --help
	Force       bool // -f принудительное копирование
	Interactive bool // -i интерактивный режим
	Verbose     bool // -v подробный вывод
}

func showHelp() {
	help := `Использование: cp [КЛЮЧИ] ИСТОЧНИК НАЗНАЧЕНИЕ
   или:  cp [ОПЦКЛЮЧИИИ] ИСТОЧНИК... ДИРЕКТОРИЯ

Ключи:
  -f            Принудительное копирование (перезапись без вопросов)
  -i            Интерактивный режим (спрашивать перед перезаписью)
  -v            Подробный вывод (показывать копируемые файлы)
  -h, --help    Показать эту справку

Описание:
  cp - копирует ИСТОЧНИК в НАЗНАЧЕНИЕ или несколько ИСТОЧНИКОВ в ДИРЕКТОРИЮ.

Примеры:
  ./cp file.txt backup.txt           Копировать файл
  ./cp -f file.txt backup.txt        Принудительная перезапись
  ./cp -i file.txt backup.txt        Спросить перед перезаписью
  ./cp -v file.txt backup.txt        Копировать с выводом
  ./cp file1 file2 file3 dir/        Копировать несколько файлов`
	fmt.Println(help)
}

func parseFlags(args []string) (CpFlags, []string, error) {
	flags := CpFlags{}
	var files []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-h" || arg == "--help" {
			flags.Help = true
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 && arg != "-" {
			for j := 1; j < len(arg); j++ {
				switch arg[j] {
				case 'f':
					flags.Force = true
					flags.Interactive = false
				case 'i':
					flags.Interactive = true
					flags.Force = false
				case 'v':
					flags.Verbose = true
				case 'h':
					flags.Help = true
				default:
					return flags, nil, fmt.Errorf("неизвестный флаг: -%c", arg[j])
				}
			}
		} else {
			files = append(files, arg)
		}
	}

	return flags, files, nil
}

func confirmOverwrite(path string) bool {
	fmt.Printf("cp: перезаписать '%s'? (y/n): ", path)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func copyFile(src, dst string, flags CpFlags) error {
	// Проверяем существование источника
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("не удаётся получить доступ к '%s': %v", src, err)
	}

	// Проверяем что это обычный файл
	if !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("'%s' не является обычным файлом", src)
	}

	// Проверяем существование назначения
	if _, err := os.Stat(dst); err == nil {
		// Файл назначения существует
		if flags.Interactive {
			if !confirmOverwrite(dst) {
				if flags.Verbose {
					fmt.Printf("cp: '%s' не перезаписан\n", dst)
				}
				return nil
			}
		} else if !flags.Force {
			return fmt.Errorf("'%s' уже существует (используй -f для перезаписи)", dst)
		}
	}

	// Открываем исходный файл
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("не удаётся открыть '%s': %v", src, err)
	}
	defer srcFile.Close()

	// Создаём файл назначения
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("не удаётся создать '%s': %v", dst, err)
	}
	defer dstFile.Close()

	// Копируем содержимое
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("ошибка копирования: %v", err)
	}

	// Копируем права доступа
	err = os.Chmod(dst, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("не удаётся установить права доступа: %v", err)
	}

	if flags.Verbose {
		fmt.Printf("'%s' -> '%s'\n", src, dst)
	}

	return nil
}

func main() {
	flags, files, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "cp: %v\n", err)
		os.Exit(1)
	}

	if flags.Help {
		showHelp()
		os.Exit(0)
	}

	if len(files) < 2 {
		fmt.Fprintf(os.Stderr, "cp: недостаточно аргументов\n")
		fmt.Fprintf(os.Stderr, "Используй 'cp --help' для справки\n")
		os.Exit(1)
	}

	// Последний аргумент назначение
	dst := files[len(files)-1]
	sources := files[:len(files)-1]

	// Проверяем назначение
	dstInfo, err := os.Stat(dst)
	isDir := err == nil && dstInfo.IsDir()

	// Если несколько источников назначение должно быть директорией
	if len(sources) > 1 && !isDir {
		fmt.Fprintf(os.Stderr, "cp: назначение '%s' не является директорией\n", dst)
		os.Exit(1)
	}

	// Копируем каждый источник
	hasError := false
	for _, src := range sources {
		// Определяем путь назначения
		var targetPath string
		if isDir {
			targetPath = filepath.Join(dst, filepath.Base(src))
		} else {
			targetPath = dst
		}

		// Копируем файл
		err = copyFile(src, targetPath, flags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cp: %v\n", err)
			hasError = true
		}
	}

	if hasError {
		os.Exit(1)
	}
}
