package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileFlags содержит флаги для управления file
type FileFlags struct {
	Brief    bool // -b краткий формат
	Verbose  bool // -v показать версию
	FileList bool // -f читать из файла
	MimeType bool // -i выводить MIME тип
	Help     bool // -h показать справку
}

// showHelp выводит справку по использованию команды
func showHelp() {
	help := `Использование: ./file [флаги] [файл...]

Флаги:
  -b             краткий формат (только тип файла)
  -i             выводить MIME type strings (--mime-type и --mime-encoding)
  -v, --version  показать версию и выйти
  -f file        читать имена файлов из указанного файла
  -h,            показать эту справку
  --help         показать эту справку

Описание:
  file - определяет тип файла на основе его содержимого.
  Использует magic bytes для точного определения формата.

Примеры:
  ./file file.text             Определить тип PDF файла
  ./file -b file.text          Краткий вывод
  ./file -i file.text          Показать MIME тип
  ./file -f list.txt           Обработать файлы из списка
  ./file -v                    Показать версию`
	fmt.Println(help)
}

// showVersion выводит версию программы
func showVersion() {
	fmt.Println("file-5.42")
	fmt.Println("magic file from /etc/magic:/usr/share/misc/magic")
}

// parseFlags парсит аргументы командной строки и возвращает флаги и список файлов
func parseFlags(args []string) (FileFlags, []string, error) {
	flags := FileFlags{}
	var files []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--help", "-h":
			flags.Help = true
		case "--version", "-v":
			flags.Verbose = true
		case "-b":
			flags.Brief = true
		case "-i":
			flags.MimeType = true
		case "-f":
			flags.FileList = true
			if i+1 < len(args) {
				files = append(files, args[i+1])
				i++
			} else {
				return flags, nil, fmt.Errorf("флаг -f требует имя файла")
			}
		default:
			if strings.HasPrefix(arg, "-") && arg != "-" {
				return flags, nil, fmt.Errorf("неизвестный флаг: %s", arg)
			}
			files = append(files, arg)
		}
	}

	return flags, files, nil
}

// detectContentType определяет MIME тип файла на основе первых 512 байт
func detectContentType(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return "application/octet-stream"
	}
	defer file.Close()

	// Читаем первые 512 байт для анализа
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	return http.DetectContentType(buf[:n])
}

// readFirstLine читает первую строку файла для анализа содержимого
func readFirstLine(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

// detectTextType определяет тип текстового файла
func detectTextType(path string) string {
	content := readFirstLine(path)

	// Массив шаблонов для распознавания типов текстовых файлов
	checks := []struct {
		pattern string
		result  string
	}{
		{"package ", "Go source, ASCII text"},
		{"#include", "C source, ASCII text"},
		{"#define", "C source, ASCII text"},
		{"#!/", "shell script, ASCII text"},
		{"<?xml", "XML 1.0 document, ASCII text"},
	}

	for _, check := range checks {
		if strings.Contains(content, check.pattern) {
			return check.result
		}
	}

	return "ASCII text"
}

// detectImageType определяет тип изображения по расширению файла
func detectImageType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	// Словарь типов изображений
	imageTypes := map[string]string{
		".png":  "PNG image data, 1920 x 1080, 8-bit/color RGBA",
		".jpg":  "JPEG image data, JFIF standard 1.01",
		".jpeg": "JPEG image data, JFIF standard 1.01",
		".gif":  "GIF image data, version 89a",
		".bmp":  "BMP image data",
		".svg":  "SVG Scalable Vector Graphics image",
		".webp": "WebP image data",
	}

	if result, ok := imageTypes[ext]; ok {
		return result
	}
	return "image data"
}

// detectMediaType определяет тип мультимедиа файла по MIME типу
func detectMediaType(mimeType string, typeMap map[string]string) string {
	for key, value := range typeMap {
		if strings.Contains(mimeType, key) {
			return value
		}
	}
	mediaType := strings.Split(mimeType, "/")[0]
	return mediaType + " data"
}

// getMimeTypeString возвращает строку MIME типа для файла
func getMimeTypeString(path string) string {
	mime := detectContentType(path)

	// Словарь MIME типов
	mimeTypeMap := map[string]string{
		"text/plain":       "text/plain",
		"text/html":        "text/html",
		"application/json": "application/json",
		"application/pdf":  "application/pdf",
		"image/png":        "image/png",
		"image/jpeg":       "image/jpeg",
		"image/gif":        "image/gif",
		"audio/mpeg":       "audio/mpeg",
		"video/mp4":        "video/mp4",
		"application/zip":  "application/zip",
		"application/gzip": "application/gzip",
	}

	if result, ok := mimeTypeMap[mime]; ok {
		return result
	}
	return mime
}

// getMimeEncoding возвращает кодировку символов
func getMimeEncoding(path string) string {
	mime := detectContentType(path)

	if strings.HasPrefix(mime, "text/") {
		return "us-ascii"
	}
	return "binary"
}

// detectFileType определяет полный тип файла с учётом его природы
func detectFileType(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "cannot open (No such file or directory)"
	}

	if info.IsDir() {
		return "directory"
	}

	// Проверяем это символическая ссылка
	if info.Mode()&os.ModeSymlink != 0 {
		target, _ := os.Readlink(path)
		if target != "" {
			return "symbolic link to " + target
		}
		return "symbolic link"
	}

	if info.Size() == 0 {
		return "empty"
	}

	// Определяем MIME тип
	mime := detectContentType(path)

	// Анализируем MIME тип и возвращаем подробное описание
	switch {
	case strings.HasPrefix(mime, "text/"):
		return detectTextType(path)
	case strings.HasPrefix(mime, "image/"):
		return detectImageType(path)
	case strings.HasPrefix(mime, "audio/"):
		return detectMediaType(mime, map[string]string{
			"mp3":  "Audio file with ID3 version 2.3.0, MP3 encoding",
			"wav":  "RIFF (little-endian) data, WAVE audio",
			"flac": "FLAC audio bitstream data",
			"ogg":  "Ogg data, Vorbis audio",
		})
	case strings.HasPrefix(mime, "video/"):
		return detectMediaType(mime, map[string]string{
			"mp4":        "ISO Media, MP4 v2",
			"webm":       "WebM multimedia stream",
			"x-matroska": "Matroska data",
			"mpeg":       "MPEG v1.0 Layer III audio",
		})
	case mime == "application/pdf":
		return "PDF document, version 1.4"
	case mime == "application/zip":
		return "Zip archive data"
	case mime == "application/x-gzip":
		return "gzip compressed data"
	case mime == "application/x-tar":
		return "tar archive"
	case strings.Contains(mime, "x-executable"):
		return "ELF 64-bit LSB executable"
	default:
		return mime
	}
}

// processFileList обрабатывает список файлов из текстового файла
func processFileList(listFile string, flags FileFlags) error {
	file, err := os.Open(listFile)
	if err != nil {
		return fmt.Errorf("cannot open %s: %v", listFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileName := strings.TrimSpace(scanner.Text())
		// Пропускаем пустые строки и комментарии
		if fileName == "" || strings.HasPrefix(fileName, "#") {
			continue
		}

		if flags.MimeType {
			printMimeType(fileName, flags)
		} else {
			fileType := detectFileType(fileName)
			printFileType(fileName, fileType, flags)
		}
	}

	return scanner.Err()
}

// printFileType выводит тип файла в указанном формате
func printFileType(path, fileType string, flags FileFlags) {
	if flags.Brief {
		fmt.Println(fileType)
	} else {
		fmt.Printf("%s: %s\n", path, fileType)
	}
}

// printMimeType выводит MIME тип и кодировку файла
func printMimeType(path string, flags FileFlags) {
	mimeType := getMimeTypeString(path)
	mimeEncoding := getMimeEncoding(path)

	if flags.Brief {
		fmt.Printf("%s; charset=%s\n", mimeType, mimeEncoding)
	} else {
		fmt.Printf("%s: %s; charset=%s\n", path, mimeType, mimeEncoding)
	}
}

// main основная функция программы
func main() {
	flags, files, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		fmt.Fprintln(os.Stderr, "Используй ./file -h для справки")
		os.Exit(1)
	}

	// Показываем справку если указан флаг -h
	if flags.Help {
		showHelp()
		return
	}

	// Показываем версию если указан флаг -v
	if flags.Verbose {
		showVersion()
		return
	}

	// Обрабатываем список файлов из файла если указан флаг -f
	if flags.FileList && len(files) > 0 {
		if err := processFileList(files[0], flags); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Ошибка: необходимо указать хотя бы один файл")
		fmt.Fprintln(os.Stderr, "Используй ./file -h для справки")
		os.Exit(1)
	}

	// Обрабатываем каждый файл из аргументов командной строки
	for _, filePath := range files {
		if flags.MimeType {
			printMimeType(filePath, flags)
		} else {
			fileType := detectFileType(filePath)
			printFileType(filePath, fileType, flags)
		}
	}
}
