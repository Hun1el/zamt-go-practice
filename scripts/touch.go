package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// TouchFlags содержит флаги для управления touch
type TouchFlags struct {
	Help       bool // -h/--help справка
	NoCreate   bool // -c/--no-create не создавать файлы
	AccessOnly bool // -a изменить только время доступа
	ModifyOnly bool // -m изменить только время изменения
	Files      []string
}

// showHelp выводит справку по использованию команды touch
func showHelp() {
	fmt.Println(`Использование: touch [КЛЮЧИ] ФАЙЛ...

Ключи:
  -a                     изменить только время доступа
  -c        			 не создавать файлы
  -m                     изменить только время изменения
  -h            		 показать эту справку
  --help                 показать эту справку

Описание:
  touch - обновляет время доступа и изменения файла.
  Если файл не существует, создаёт его (если не указан -c).

Примеры:
  ./touch file.txt                 Создать файл или обновить время
  ./touch -c file.txt              Не создавать, только обновить время
  ./touch -a file.txt              Обновить только время доступа
  ./touch -m file.txt              Обновить только время изменения`)
}

// parseFlags парсит флаги для touch
func parseFlags(args []string) (TouchFlags, error) {
	flags := TouchFlags{}
	files := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") && arg != "-" {
			if arg == "--help" || arg == "-h" {
				flags.Help = true
			} else if arg == "-a" {
				flags.AccessOnly = true
			} else if arg == "-m" {
				flags.ModifyOnly = true
			} else if arg == "-c" {
				flags.NoCreate = true
			} else if len(arg) > 2 {
				for _, ch := range arg[1:] {
					switch ch {
					case 'a':
						flags.AccessOnly = true
					case 'm':
						flags.ModifyOnly = true
					case 'c':
						flags.NoCreate = true
					case 'h':
						flags.Help = true
					default:
						return flags, fmt.Errorf("неизвестный флаг: -%c", ch)
					}
				}
			} else {
				return flags, fmt.Errorf("неизвестный флаг: %s", arg)
			}
		} else {
			files = append(files, arg)
		}
	}

	flags.Files = files
	return flags, nil
}

func main() {
	flags, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "touch: %v\n", err)
		os.Exit(1)
	}

	if flags.Help {
		showHelp()
		return
	}

	if len(flags.Files) == 0 {
		fmt.Fprintf(os.Stderr, "touch: отсутствуют аргументы файла\n")
		os.Exit(1)
	}

	now := time.Now()

	// Обрабатываем каждый файл
	for _, file := range flags.Files {
		// Проверяем существует ли файл
		info, err := os.Stat(file)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "touch: не могу получить доступ к '%s': %v\n", file, err)
				continue
			}
			// Файл не существует
			if flags.NoCreate {
				continue
			}
			// Создаём файл
			f, err := os.Create(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "touch: не могу создать '%s': %v\n", file, err)
				continue
			}
			f.Close()
			continue
		}

		// Определяем времена доступа и изменения
		atime := info.ModTime()
		mtime := info.ModTime()

		if flags.AccessOnly {
			atime = now
		} else if flags.ModifyOnly {
			mtime = now
		} else {
			// Если не указано ни -a ни -m
			atime = now
			mtime = now
		}

		// Применяем новые времена
		err = os.Chtimes(file, atime, mtime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "touch: не могу изменить время '%s': %v\n", file, err)
		}
	}
}
