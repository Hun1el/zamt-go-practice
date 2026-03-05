package main

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TarFlags содержит флаги управления tar
type TarFlags struct {
	Create  bool   // -c создать архив
	Extract bool   // -x извлечь архив
	List    bool   // -t показать содержимое
	File    string // -f имя архивного файла
	Help    bool   // -h справка
}

// showHelp выводит справку по использованию
func showHelp() {
	fmt.Println(`Использование: ./tar [флаги] архив.tar [файлы...]

Флаги:
  -c         создать архив
  -x         извлечь архив
  -t         показать содержимое
  -f FILE    имя архивного файла
  -h         показать эту справку
  --help     показать эту справку

Примеры:
  ./tar -c -f arch.tar file1.txt file2.txt
  ./tar -cf arch.tar dir/
  ./tar -x -f arch.tar
  ./tar -t -f arch.tar
  ./tar -h`)
}

// parseFlags парсит флаги из аргументов командной строки
func parseFlags(args []string) (TarFlags, []string, error) {
	flags := TarFlags{}
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") && arg != "-" {
			if arg == "--help" {
				flags.Help = true
				continue
			}

			for j := 1; j < len(arg); j++ {
				ch := arg[j]
				switch ch {
				case 'c':
					flags.Create = true
				case 'x':
					flags.Extract = true
				case 't':
					flags.List = true
				case 'f':
					if j+1 < len(arg) {
						flags.File = arg[j+1:]
						j = len(arg) - 1
					} else if i+1 < len(args) {
						i++
						flags.File = args[i]
					}
				case 'h':
					flags.Help = true
				}
			}
		} else {
			remaining = append(remaining, arg)
		}
	}

	return flags, remaining, nil
}

// addFileToTar добавляет файл в архив
func addFileToTar(tw *tar.Writer, path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tar: %s: %v\n", path, err)
		return false
	}

	if info.IsDir() {
		return addDirToTar(tw, path)
	}

	file, _ := os.Open(path)
	defer file.Close()

	header, _ := tar.FileInfoHeader(info, "")
	header.Name = path

	tw.WriteHeader(header)
	io.Copy(tw, file)

	return true
}

// addDirToTar рекурсивно добавляет директорию в архив
func addDirToTar(tw *tar.Writer, dirPath string) bool {
	entries, _ := os.ReadDir(dirPath)
	success := true

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			info, _ := entry.Info()
			header, _ := tar.FileInfoHeader(info, "")
			header.Name = fullPath + "/"
			tw.WriteHeader(header)
			if !addDirToTar(tw, fullPath) {
				success = false
			}
		} else {
			if !addFileToTar(tw, fullPath) {
				success = false
			}
		}
	}

	return success
}

// createTar создаёт архив с указанными файлами
func createTar(archiveName string, files []string) int {
	file, err := os.Create(archiveName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tar: %s: %v\n", archiveName, err)
		return 1
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	if len(files) == 0 {
		files = []string{"."}
	}

	hasError := false
	for _, filePath := range files {
		if !addFileToTar(tw, filePath) {
			hasError = true
		}
	}

	if hasError {
		fmt.Fprintf(os.Stderr, "tar: Завершение работы с состоянием неисправности из-за возникших ошибок\n")
		return 2
	}

	return 0
}

// extractTar распаковывает архив
func extractTar(archiveName string) int {
	file, err := os.Open(archiveName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tar: %s: %v\n", archiveName, err)
		return 1
	}
	defer file.Close()

	tr := tar.NewReader(file)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			return 1
		}

		// Используем switch для обработки типов файлов
		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(header.Name, 0755)
		case tar.TypeReg:
			dir := filepath.Dir(header.Name)
			os.MkdirAll(dir, 0755)

			f, _ := os.OpenFile(header.Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			io.Copy(f, tr)
			f.Close()
		}
	}

	return 0
}

// listTar показывает содержимое архива
func listTar(archiveName string) int {
	file, err := os.Open(archiveName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tar: %s: %v\n", archiveName, err)
		return 1
	}
	defer file.Close()

	tr := tar.NewReader(file)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			return 1
		}

		fmt.Println(header.Name)
	}

	return 0
}

func main() {
	flags, files, _ := parseFlags(os.Args[1:])

	if flags.Help {
		showHelp()
		return
	}

	if flags.File == "" {
		fmt.Fprintln(os.Stderr, "tar: укажите архив с -f")
		os.Exit(1)
	}

	var exitCode int

	if flags.List {
		exitCode = listTar(flags.File)
	} else if flags.Extract {
		exitCode = extractTar(flags.File)
	} else if flags.Create {
		exitCode = createTar(flags.File, files)
	} else {
		fmt.Fprintln(os.Stderr, "tar: укажите -c, -x или -t")
		exitCode = 1
	}

	os.Exit(exitCode)
}
