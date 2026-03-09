package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
)

// ArchOptions содержит флаги для управления arch
type ArchOptions struct {
	Help    bool
	Version bool
}

// Функция для вывода справки
func printHelpArch() {
	helpText := `Использование: arch [ОПЦИИ]
  
Ключи:
  -h      		      Показать эту справку
   --help             Показать эту справку
  -v, --version       Показать версию

Описание:
  arch - выводит архитектуру процессора системы.
            
Примеры:
  ./arch              Показать архитектуру процессора
  ./arch -h           Показать эту справку
  ./arch --version    Показать версию`
	fmt.Println(helpText)
}

// Функция для вывода версии
func printVersionArch() {
	versionText := `arch (GNU coreutils) 9.1
Copyright (C) 2022 Free Software Foundation, Inc.
Лицензия GPLv3+: GNU GPL версии 3 или новее <https://gnu.org/licenses/gpl.html>.
Это свободное ПО: вы можете изменять и распространять его.
Нет НИКАКИХ ГАРАНТИЙ в пределах действующего законодательства.

Авторы программы — David MacKenzie и Karel Zak.`
	fmt.Println(versionText)
}

// Функция для преобразования массива int8 в строку
func bytesToString(b [65]int8) string {
	bytes := make([]byte, len(b))
	for i, v := range b {
		bytes[i] = byte(v)
	}
	return strings.TrimRight(string(bytes), "\x00")
}

// Функция для получения архитектуры процессора
func getArchitecture() (string, error) {
	var uts syscall.Utsname
	err := syscall.Uname(&uts)
	if err != nil {
		return "", fmt.Errorf("ошибка получения информации системы: %v", err)
	}

	return bytesToString(uts.Machine), nil
}

// Функция для парсинга флагов
func parseArgumentsArch() (*ArchOptions, error) {
	options := &ArchOptions{}

	flagSet := flag.NewFlagSet("arch", flag.ContinueOnError)
	flagSet.SetOutput(os.Stdout)

	flagSet.BoolVar(&options.Help, "h", false, "Показать справку")
	flagSet.BoolVar(&options.Help, "help", false, "Показать справку")
	flagSet.BoolVar(&options.Version, "v", false, "Показать версию")
	flagSet.BoolVar(&options.Version, "version", false, "Показать версию")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("ошибка разбора аргументов: %v", err)
	}

	if flagSet.NArg() > 0 {
		return nil, fmt.Errorf("неожиданные аргументы: %v\nИспользуйте --help для справки", flagSet.Args())
	}

	return options, nil
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Критическая ошибка (panic): %v\n", r)
			os.Exit(1)
		}
	}()

	options, err := parseArgumentsArch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}

	if options.Help {
		printHelpArch()
		os.Exit(0)
	}

	if options.Version {
		printVersionArch()
		os.Exit(0)
	}

	arch, err := getArchitecture()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка получения архитектуры: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(arch)

	os.Exit(0)
}
