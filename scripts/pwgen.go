package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
)

const (
	DEFAULT_LENGTH = 8
	DEFAULT_COUNT  = 160
)

var (
	consonants = "bcdfghjklmnpqrstvwxyz"
	vowels     = "aeiou"
	capitals   = "BCDFGHJKLMNPQRSTVWXYZ"
	digits     = "0123456789"
	symbols    = "!@#$%^&*-_+=[]{}|:;<>,.?/"
)

// PwgenFlags содержит флаги для управления генерацией пароля
type PwgenFlags struct {
	NoCapital  bool // -A НЕ включать заглавные буквы
	NoNumerals bool // -0 НЕ включать цифры
	Symbols    bool // -y включить символы
	RequireNum bool // -n требовать хотя бы одну цифру
	Help       bool // -h/--help справка
}

// ShowHelp выводит справку по использованию команды pwgen
func ShowHelp() {
	help := `
Использование: pwgen [ФЛАГИ] [длина] [количество]

Флаги:
  -A         НЕ включать заглавные буквы
  -0         НЕ включать цифры
  -y         Включить символы (!@#$%^&*)
  -n         Требовать хотя бы одну цифру в пароле
  -h         Показать эту справку
  --help     Показать эту справку

Описание:
  pwgen - генератор безопасных паролей.

Примеры:
  ./pwgen              Пароли длины 8, 160 штук
  ./pwgen 12           Пароли длины 12, 160 штук
  ./pwgen 10 5         5 паролей длины 10
  ./pwgen -A 12        Без заглавных букв
  ./pwgen -n 12        С цифрами
  ./pwgen -0 12        Без цифр
  ./pwgen -y 12        С символами
`
	fmt.Println(help)
}

// ParsePwgenFlags парсит флаги для команды pwgen
func ParsePwgenFlags(args []string) (PwgenFlags, []string, error) {
	flags := PwgenFlags{}
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			flags.Help = true
		} else if strings.HasPrefix(arg, "-") && arg != "-" {
			// Обработка комбинированных флагов
			for j := 1; j < len(arg); j++ {
				ch := rune(arg[j])
				switch ch {
				case 'A':
					flags.NoCapital = true
				case '0':
					flags.NoNumerals = true
				case 'y':
					flags.Symbols = true
				case 'n':
					flags.RequireNum = true
				case 'h':
					flags.Help = true
				default:
					return flags, nil, fmt.Errorf("неизвестный флаг: -%c", ch)
				}
			}
		} else {
			remaining = append(remaining, arg)
		}
	}

	return flags, remaining, nil
}

// BuildCharacterSet создаёт набор символов для пароля
func BuildCharacterSet(flags PwgenFlags) string {
	charset := "abcdefghijklmnopqrstuvwxyz"

	// По умолчанию добавляем ЗАГЛАВНЫЕ
	if !flags.NoCapital {
		charset += capitals
	}

	// По умолчанию добавляем ЦИФРЫ
	if !flags.NoNumerals {
		charset += digits
	}

	// Добавляем символы только если флаг -y
	if flags.Symbols {
		charset += symbols
	}

	return charset
}

// GeneratePassword генирирует один пароль
func GeneratePassword(length int, charset string, flags PwgenFlags) string {
	password := make([]rune, length)

	for i := 0; i < length; i++ {
		idx := randInt(len(charset))
		password[i] = rune(charset[idx])
	}

	// Если требуется цифра заменяем случайный символ
	if flags.RequireNum && !flags.NoNumerals {
		hasNum := false
		for _, ch := range password {
			if ch >= '0' && ch <= '9' {
				hasNum = true
				break
			}
		}
		if !hasNum {
			pos := randInt(length)
			password[pos] = rune(digits[randInt(len(digits))])
		}
	}

	return string(password)
}

// randInt возвращает случайное число от 0 до max-1
func randInt(max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

// ParseArguments парсит аргументы длины и количества
func ParseArguments(args []string) (int, int, error) {
	length := DEFAULT_LENGTH
	count := DEFAULT_COUNT

	if len(args) >= 1 {
		var err error
		length, err = strconv.Atoi(args[0])
		if err != nil || length <= 0 {
			return 0, 0, fmt.Errorf("'%s' не является положительным числом", args[0])
		}
	}

	if len(args) >= 2 {
		var err error
		count, err = strconv.Atoi(args[1])
		if err != nil || count <= 0 {
			return 0, 0, fmt.Errorf("'%s' не является положительным числом", args[1])
		}
	}

	if len(args) > 2 {
		return 0, 0, fmt.Errorf("слишком много аргументов")
	}

	return length, count, nil
}

func Print8x20Grid(passwords []string) {
	for i, password := range passwords {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(password)

		if (i+1)%8 == 0 {
			fmt.Println()
		}
	}
	if len(passwords)%8 != 0 {
		fmt.Println()
	}
}

func main() {
	flags, args, err := ParsePwgenFlags(os.Args[1:])
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("Используй pwgen --help для справки")
		os.Exit(1)
	}

	if flags.Help {
		ShowHelp()
		return
	}

	// Парсим аргументы длины и количества
	length, count, err := ParseArguments(args)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Println("Используй pwgen --help для справки")
		os.Exit(1)
	}

	// Строим набор символов
	charset := BuildCharacterSet(flags)
	if charset == "" {
		fmt.Println("Ошибка: набор символов пуст")
		os.Exit(1)
	}

	// Генерируем пароли
	passwords := make([]string, 0, count)
	for i := 0; i < count; i++ {
		password := GeneratePassword(length, charset, flags)
		passwords = append(passwords, password)
	}

	// Выводим пароли сеткой
	Print8x20Grid(passwords)
}
