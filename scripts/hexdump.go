package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// HexdumpFlags содержит флаги для управления выводом hexdump
type HexdumpFlags struct {
	Canonical bool   // -C канонический формат
	Length    uint64 // -n читать N байт
	Skip      uint64 // -s пропустить N байт
	Help      bool   // -h/--help справка
}

// ShowHelp выводит справку по использованию команды hexdump
func ShowHelp() {
	help := `
Использование: hexdump [флаги] [файл...]

Кл:
  -C         Канонический формат (hex + ASCII)
  -n N       Читать N байт
  -s N       Пропустить N байт
  -h         Показать эту справку
  --help     Показать эту справку

Описание:
  hexdump - выводит содержимое файла в шестнадцатеричном формате.

Примеры:
  hexdump file.bin               Стандартный формат
  hexdump -C file.bin            Канонический формат
  hexdump -s 100 file.bin        Пропустить первые 100 байт
  hexdump -n 512 file.bin        Читать только 512 байт
  hexdump -Cs 100 -n 256 file    Комбинированные флаги
`
	fmt.Println(help)
}

// ParseHexdumpFlags парсит флаги для команды hexdump
func ParseHexdumpFlags(args []string) (HexdumpFlags, []string, error) {
	flags := HexdumpFlags{}
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			flags.Help = true
		} else if strings.HasPrefix(arg, "-") && arg != "-" {
			// Обработка флагов
			j := 1
			for j < len(arg) {
				ch := rune(arg[j])
				switch ch {
				case 'C':
					flags.Canonical = true
					j++
				case 'n':
					if j+1 < len(arg) {
						// Формат -nNNN
						_, err := fmt.Sscanf(arg[j+1:], "%d", &flags.Length)
						if err != nil {
							return flags, nil, fmt.Errorf("неверный аргумент для -n: %s", arg[j+1:])
						}
						return flags, remaining, nil
					} else if i+1 < len(args) {
						// Формат -n NNN
						_, err := fmt.Sscanf(args[i+1], "%d", &flags.Length)
						if err != nil {
							return flags, nil, fmt.Errorf("неверный аргумент для -n: %s", args[i+1])
						}
						i++
						j++
					} else {
						return flags, nil, fmt.Errorf("флаг -n требует аргумента")
					}
				case 's':
					// -s N - может быть -sN или -s N
					if j+1 < len(arg) {
						// Формат -sNNN
						_, err := fmt.Sscanf(arg[j+1:], "%d", &flags.Skip)
						if err != nil {
							return flags, nil, fmt.Errorf("неверный аргумент для -s: %s", arg[j+1:])
						}
						return flags, remaining, nil
					} else if i+1 < len(args) {
						// Формат -s NNN
						_, err := fmt.Sscanf(args[i+1], "%d", &flags.Skip)
						if err != nil {
							return flags, nil, fmt.Errorf("неверный аргумент для -s: %s", args[i+1])
						}
						i++
						j++
					} else {
						return flags, nil, fmt.Errorf("флаг -s требует аргумента")
					}
				case 'h':
					flags.Help = true
					j++
				default:
					return flags, nil, fmt.Errorf("неизвестный флаг: -%c", ch)
				}
			}
		} else {
			// Это файл
			remaining = append(remaining, arg)
		}
	}

	return flags, remaining, nil
}

// HexdumpFile выводит содержимое файла в hex формате
func HexdumpFile(filename string, flags HexdumpFlags) error {
	var f *os.File
	var err error
	var fileSize int64

	// Открываем файл или читаем из stdin
	if filename == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		// Узнаём реальный размер файла
		fi, err := f.Stat()
		if err == nil {
			fileSize = fi.Size()
		}
	}

	// Пропускаем байты если нужно
	if flags.Skip > 0 {
		_, err := f.Seek(int64(flags.Skip), 0)
		if err != nil {
			return err
		}
	}

	// Читаем весь файл в буфер
	var allData []byte
	buf := make([]byte, 16)

	for {
		n, err := f.Read(buf)
		if n == 0 {
			break
		}
		if err != nil && err != io.EOF {
			return err
		}

		allData = append(allData, buf[:n]...)

		// Проверяем лимит длины
		if flags.Length > 0 && uint64(len(allData)) >= flags.Length {
			allData = allData[:flags.Length]
			break
		}
	}

	// Теперь выводим весь прочитанный буфер
	for i := 0; i < len(allData); i += 16 {
		end := i + 16
		if end > len(allData) {
			end = len(allData)
		}

		// Смещение в файле (реальное)
		fileOffset := flags.Skip + uint64(i)
		printLine(allData[i:end], flags.Canonical, fileOffset)
	}

	// Выводим финальное смещение
	var finalOffset uint64
	if fileSize > 0 {
		finalOffset = uint64(fileSize)
	} else {
		finalOffset = flags.Skip + uint64(len(allData))
	}
	fmt.Printf("%07x\n", finalOffset)
	return nil
}

// printLine выводит одну строку в hex формате
func printLine(data []byte, canonical bool, byteOffset uint64) {
	if canonical {
		// Канонический формат offset + hex байты + ASCII
		fmt.Printf("%08x  ", byteOffset)

		// Выводим hex байты
		for i := 0; i < len(data); i++ {
			fmt.Printf("%02x ", data[i])
		}

		// Выравниваем до 16 байт
		for i := len(data); i < 16; i++ {
			fmt.Print("   ")
		}

		fmt.Print(" |")

		// Выводим ASCII
		for _, b := range data {
			if b >= 32 && b <= 126 {
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".")
			}
		}

		// Выравниваем ASCII до 16 символов
		for i := len(data); i < 16; i++ {
			fmt.Print(" ")
		}

		fmt.Println("|")
	} else {
		// Стандартный формат: выводим смещение как есть
		fmt.Printf("%07x ", byteOffset)

		for i := 0; i < len(data); i += 2 {
			if i+1 < len(data) {
				val := uint16(data[i+1])<<8 | uint16(data[i])
				fmt.Printf("%04x ", val)
			} else {
				fmt.Printf("%02x ", data[i])
			}
		}

		fmt.Println()
	}
}

func main() {
	flags, files, err := ParseHexdumpFlags(os.Args[1:])
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("Используй hexdump --help для справки")
		os.Exit(1)
	}

	if flags.Help {
		ShowHelp()
		return
	}

	// Если файлы не указаны, читаем из stdin
	if len(files) == 0 {
		files = []string{"-"}
	}

	// Обрабатываем каждый файл
	for _, filename := range files {
		if err := HexdumpFile(filename, flags); err != nil {
			fmt.Printf("hexdump: %s: %v\n", filename, err)
		}
	}
}
