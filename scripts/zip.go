package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipFlags содержит флаги управления zip
type ZipFlags struct {
	Recursive bool // -r рекурсивно добавлять директории
	Delete    bool // -d удалить файлы из архива
	Quiet     bool // -q тихий режим
	Help      bool // -h справка
}

// showHelp выводит справку по использованию
func showHelp() {
	fmt.Println(`Использование: ./zip [КЛЮЧИ] arch.zip [файлы...]

Ключи:
  -r         рекурсивно добавлять директории
  -d         удалить указанные файлы из архива
  -q         тихий режим
  -h         показать эту справку
  --help     показать эту справку

Описание:
  zip - создаёт или модифицирует ZIP архивы.
  
  Добавляет файлы в архив или удаляет их из архива.
  С флагом -r добавляет директории рекурсивно.

Примеры:
  ./zip arch.zip file1.txt file2.txt      Добавить файлы
  ./zip -r arch.zip test/                 Добавить директорию рекурсивно
  ./zip -d arch.zip file.txt              Удалить файл из архива
  ./zip -q arch.zip file.txt              Добавить в тихом режиме
  ./zip -rq arch.zip test/                Рекурсивно и тихо`)
}

// parseFlags парсит флаги из аргументов командной строки
func parseFlags(args []string) (ZipFlags, string, []string, error) {
	flags := ZipFlags{}
	var files []string
	archiveName := ""

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && arg != "-" {
			if arg == "--help" || arg == "-h" {
				flags.Help = true
			} else if len(arg) > 2 {
				for _, ch := range arg[1:] {
					switch ch {
					case 'r':
						flags.Recursive = true
					case 'd':
						flags.Delete = true
					case 'q':
						flags.Quiet = true
					case 'h':
						flags.Help = true
					default:
						return flags, "", nil, fmt.Errorf("неизвестный флаг: -%c", ch)
					}
				}
			} else {
				switch arg {
				case "-r":
					flags.Recursive = true
				case "-d":
					flags.Delete = true
				case "-q":
					flags.Quiet = true
				default:
					return flags, "", nil, fmt.Errorf("неизвестный флаг: %s", arg)
				}
			}
		} else {
			if archiveName == "" {
				archiveName = arg
			} else {
				files = append(files, arg)
			}
		}
	}

	return flags, archiveName, files, nil
}

// addFile добавляет файл в архив
func addFile(w *zip.Writer, path, base string, quiet bool) error {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return err
	}

	header, _ := zip.FileInfoHeader(info)
	if base != "" {
		rel, _ := filepath.Rel(base, path)
		header.Name = filepath.ToSlash(rel)
	} else {
		header.Name = filepath.Base(path)
	}
	header.Method = zip.Deflate

	writer, _ := w.CreateHeader(header)
	file, _ := os.Open(path)
	defer file.Close()
	io.Copy(writer, file)

	if !quiet {
		fmt.Printf("  adding: %s (stored 0%%)\n", header.Name)
	}
	return nil
}

// addDir добавляет директорию рекурсивно в архив
func addDir(w *zip.Writer, path, archiveName string, quiet bool) error {
	clean := strings.TrimSuffix(path, "/")
	name := filepath.Base(clean)
	absArch, _ := filepath.Abs(archiveName)

	// Добавляем саму директорию первой
	header := &zip.FileHeader{
		Name:   filepath.ToSlash(name) + "/",
		Method: zip.Store,
	}
	w.CreateHeader(header)
	if !quiet {
		fmt.Printf("  adding: %s/ (stored 0%%)\n", name)
	}

	return filepath.Walk(clean, func(p string, info os.FileInfo, err error) error {
		if err != nil || p == clean {
			return err
		}

		// Исключаем архив из архивирования
		abs, _ := filepath.Abs(p)
		if abs == absArch {
			return nil
		}

		rel, _ := filepath.Rel(clean, p)

		if info.IsDir() {
			header := &zip.FileHeader{
				Name:   filepath.ToSlash(name+"/"+rel) + "/",
				Method: zip.Store,
			}
			w.CreateHeader(header)
		} else {
			header, _ := zip.FileInfoHeader(info)
			header.Name = filepath.ToSlash(name + "/" + rel)
			header.Method = zip.Deflate

			writer, _ := w.CreateHeader(header)
			file, _ := os.Open(p)
			io.Copy(writer, file)
			file.Close()

			if !quiet {
				fmt.Printf("  adding: %s (stored 0%%)\n", header.Name)
			}
		}
		return nil
	})
}

// createZip создаёт новый архив с файлами
func createZip(archiveName string, files []string, flags ZipFlags) error {
	if !flags.Quiet {
		fmt.Printf("  adding: %s\n", filepath.Base(archiveName))
	}

	tmp := archiveName + ".tmp"
	f, _ := os.Create(tmp)
	w := zip.NewWriter(f)

	for _, file := range files {
		info, _ := os.Stat(file)
		if info.IsDir() {
			if flags.Recursive {
				addDir(w, file, archiveName, flags.Quiet)
			} else {
				w.Close()
				f.Close()
				os.Remove(tmp)
				return fmt.Errorf("%s is a directory (use -r to recurse)", file)
			}
		} else {
			addFile(w, file, "", flags.Quiet)
		}
	}

	w.Close()
	f.Close()
	os.Remove(archiveName)
	os.Rename(tmp, archiveName)

	if !flags.Quiet {
		fmt.Println("  zip complete")
	}
	return nil
}

// deleteArchive удаляет файлы из архива
func deleteArchive(archiveName string, files []string, quiet bool) error {
	r, err := zip.OpenReader(archiveName)
	if err != nil {
		return fmt.Errorf("cannot open %s: %v", archiveName, err)
	}
	defer r.Close()

	del := make(map[string]bool)
	for _, f := range files {
		del[filepath.Base(f)] = true
	}

	tmp := archiveName + ".tmp"
	out, _ := os.Create(tmp)
	w := zip.NewWriter(out)

	for _, file := range r.File {
		if del[filepath.Base(file.Name)] {
			if !quiet {
				fmt.Printf("  deleting: %s\n", file.Name)
			}
			continue
		}

		rc, _ := file.Open()
		header, _ := zip.FileInfoHeader(file.FileInfo())
		header.Name = file.Name
		header.Method = file.Method

		wc, _ := w.CreateHeader(header)
		io.Copy(wc, rc)
		rc.Close()

		if !quiet {
			fmt.Printf("  keeping: %s\n", file.Name)
		}
	}

	w.Close()
	out.Close()
	os.Remove(archiveName)
	os.Rename(tmp, archiveName)
	return nil
}

func main() {
	flags, archiveName, files, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}

	if flags.Help {
		showHelp()
		return
	}

	if archiveName == "" || len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Ошибка: необходимо указать архив и файлы")
		os.Exit(1)
	}

	var e error
	if flags.Delete {
		e = deleteArchive(archiveName, files, flags.Quiet)
	} else {
		e = createZip(archiveName, files, flags)
	}

	if e != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", e)
		os.Exit(1)
	}
}
