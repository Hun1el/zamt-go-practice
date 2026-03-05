package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// MkdirFlags содержит флаги для управления созданием директорий
type MkdirFlags struct {
	Parents bool   // -p создавать родительские директории если нужно
	Mode    string // -m устанавливать права доступа (например 755, 777)
	Verbose bool   // -v подробный вывод информации
	Help    bool   // -h/--help справка
}

// ShowHelp выводит справку по использованию команды mkdir
func ShowHelp() {
	help := `
Использование: ./mkdir [КЛЮЧИ] путь [путь2] [путь3] ...

Ключи:
  -p         Создавать родительские директории если нужно (как -p в реальном mkdir)
  -m         Устанавливать права доступа для создаваемой директории (755, 777 и т.д.)
  -v         Подробный вывод информации о создаваемых директориях
  -h         Показать эту справку
  --help     Показать эту справку

Описание:
  mkdir - создаёт одну или несколько директорий.
  
  Команда mkdir используется для создания новых папок (директорий).

Права доступа:
  7 = rwx (чтение, запись, выполнение)
  6 = rw- (чтение, запись)
  5 = r-x (чтение, выполнение)
  4 = r-- (только чтение)
  0 = --- (без прав)

Примеры:
  ./mkdir Doc                     Создать директорию Documents
  ./mkdir -p /home/user/tt   	  Создать директорию со всеми родительскими
  ./mkdir -m 777 public           Создать директорию с правами 777
  ./mkdir -pm 755 src/main        Создать со всеми родительскими и правами 755
  ./mkdir -p -m 755 src/main      Создать со всеми родительскими и правами 755
  ./mkdir -v folder1 folder2      Создать с подробным выводом информации
  ./mkdir -pv -m 755 src/test     Создать со всеми опциями и информацией
  ./mkdir folder1 folder2         Создать несколько директорий
`
	fmt.Println(help)
}

// ParseMkdirFlags парсит флаги для mkdir
func ParseMkdirFlags(args []string) (MkdirFlags, []string, error) {
	flags := MkdirFlags{}
	var paths []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			flags.Help = true
		} else if strings.HasPrefix(arg, "-") && arg != "-" {
			// Обработка комбинированных флагов
			for j := 1; j < len(arg); j++ {
				ch := rune(arg[j])
				switch ch {
				case 'p':
					flags.Parents = true
				case 'v':
					flags.Verbose = true
				case 'm':
					// Флаг -m требует значение
					// Проверяем есть ли значение в этом же аргументе
					if j+1 < len(arg) {
						// Значение в том же аргументе
						flags.Mode = arg[j+1:]
						j = len(arg) // Выходим из цикла
					} else if i+1 < len(args) {
						// Значение в следующем аргументе
						flags.Mode = args[i+1]
						i++ // Пропускаем следующий аргумент
					} else {
						return flags, nil, fmt.Errorf("ошибка: флаг -m требует значение (права доступа)")
					}
				case 'h':
					flags.Help = true
				default:
					return flags, nil, fmt.Errorf("неизвестный флаг: -%c", ch)
				}
			}
		} else {
			paths = append(paths, arg)
		}
	}

	return flags, paths, nil
}

// ParseOctal преобразует восьмеричное число в десятичное
func ParseOctal(octalStr string) (os.FileMode, error) {
	// Преобразуем восьмеричную строку в число
	mode64, err := strconv.ParseInt(octalStr, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("ошибка: неверные права доступа '%s' (используй восьмеричный формат: 755, 777)", octalStr)
	}

	return os.FileMode(mode64), nil
}

// ValidatePath проверяет корректность пути
func ValidatePath(path string) error {
	// Проверяем что путь не пустой
	if path == "" {
		return fmt.Errorf("ошибка: путь не может быть пустым")
	}

	// Проверяем что путь не содержит недопустимых символов
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("ошибка: путь содержит недопустимые символы")
	}

	return nil
}

// CreateDirectory создаёт директорию
func CreateDirectory(path string, mode os.FileMode, parents bool, verbose bool) error {
	// Проверяем что директория не существует
	info, err := os.Stat(path)
	if err == nil {
		// Директория существует
		if info.IsDir() {
			if verbose {
				fmt.Printf("mkdir: директория '%s' уже существует\n", path)
			}
			return fmt.Errorf("ошибка: директория '%s' уже существует", path)
		}
		return fmt.Errorf("ошибка: '%s' не является директорией", path)
	}

	if parents {
		// Создаём все необходимые родительские директории
		err = os.MkdirAll(path, mode)
	} else {
		// Создаём только одну директорию
		err = os.Mkdir(path, mode)
	}

	if err != nil {
		return fmt.Errorf("ошибка при создании директории '%s': %v", path, err)
	}

	if verbose {
		fmt.Printf("mkdir: директория '%s' создана с правами %04o\n", path, mode)
	}

	return nil
}

func main() {
	flags, paths, err := ParseMkdirFlags(os.Args[1:])
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("Используй ./mkdir --help для справки")
		os.Exit(1)
	}

	if flags.Help {
		ShowHelp()
		return
	}

	// Проверяем что указаны пути
	if len(paths) == 0 {
		fmt.Printf("Ошибка: необходимо указать хотя бы один путь\n")
		fmt.Println("Используй ./mkdir --help для справки")
		os.Exit(1)
	}

	var mode os.FileMode = 0755 // По умолчанию 755
	if flags.Mode != "" {
		var err error
		mode, err = ParseOctal(flags.Mode)
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	}

	// Создаём все директории
	exitCode := 0
	for _, path := range paths {
		// Проверяем валидность пути
		err := ValidatePath(path)
		if err != nil {
			fmt.Printf("%v\n", err)
			exitCode = 1
			continue
		}

		// Создаём директорию
		err = CreateDirectory(path, mode, flags.Parents, flags.Verbose)
		if err != nil {
			fmt.Printf("%v\n", err)
			exitCode = 1
			continue
		}
	}

	os.Exit(exitCode)
}
