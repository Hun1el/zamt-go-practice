package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
)

// UnameOptions содержит флаги для управления uname
type UnameOptions struct {
	Help      bool
	System    bool // -s имя системы
	Nodename  bool // -n имя узла
	Processor bool // -p архитектура процессора
	All       bool // -a вся информация
}

// UnameInfo содержит информацию о системе
type UnameInfo struct {
	System    string
	Nodename  string
	Release   string
	Version   string
	Machine   string
	Processor string
}

// Функция для вывода справки
func printHelpUname() {
	helpText := `Использование: uname [ОПЦИИ]
  
Ключи:
  -s            Имя системы (ОС)
  -n            Имя узла (hostname)
  -p            Архитектура процессора
  -a            Вся информация
  -h    	    Показать эту справку
  --help 		Показать эту справку

  Описание:
  uname - выводит информацию о системе.
            
Примеры:
  ./uname -s        Показать имя системы
  ./uname -n        Показать имя узла
  ./uname -p        Показать архитектуру процессора
  ./uname -a        Показать всю информацию`
	fmt.Println(helpText)
}

// Функция для преобразования массива int8 в строку
func bytesToString(b [65]int8) string {
	bytes := make([]byte, len(b))
	for i, v := range b {
		bytes[i] = byte(v)
	}
	// Удаляем null-терминаторы
	return strings.TrimRight(string(bytes), "\x00")
}

// Функция для получения информации о системе через syscall
func getUnameInfo() (*UnameInfo, error) {
	info := &UnameInfo{}

	var uts syscall.Utsname
	err := syscall.Uname(&uts)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации системы: %v", err)
	}

	// Преобразуем массивы int8 в строки
	info.System = bytesToString(uts.Sysname)
	info.Nodename = bytesToString(uts.Nodename)
	info.Release = bytesToString(uts.Release)
	info.Version = bytesToString(uts.Version)
	info.Machine = bytesToString(uts.Machine)

	// Для процессора используем машину
	info.Processor = info.Machine

	return info, nil
}

// Функция для парсинга флагов
func parseArgumentsUname() (*UnameOptions, error) {
	options := &UnameOptions{}

	flagSet := flag.NewFlagSet("uname", flag.ContinueOnError)
	flagSet.SetOutput(os.Stdout)

	flagSet.BoolVar(&options.Help, "h", false, "Показать справку")
	flagSet.BoolVar(&options.Help, "help", false, "Показать справку")
	flagSet.BoolVar(&options.System, "s", false, "Имя системы")
	flagSet.BoolVar(&options.Nodename, "n", false, "Имя узла")
	flagSet.BoolVar(&options.Processor, "p", false, "Архитектура процессора")
	flagSet.BoolVar(&options.All, "a", false, "Вся информация")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("ошибка разбора аргументов: %v", err)
	}

	if flagSet.NArg() > 0 {
		return nil, fmt.Errorf("неожиданные аргументы: %v\nИспользуйте -h для справки", flagSet.Args())
	}

	return options, nil
}

// Функция для вывода информации согласно флагам
func printUnameInfo(info *UnameInfo, options *UnameOptions) {
	// Если не указано никаких флагов выводим только имя системы
	if !options.System && !options.Nodename && !options.Processor && !options.All {
		fmt.Println(info.System)
		return
	}

	var output []string

	// Если -a выводим всё
	if options.All {
		output = append(output, info.System)
		output = append(output, info.Nodename)
		output = append(output, info.Release)
		output = append(output, info.Version)
		output = append(output, info.Machine)
		fmt.Println(strings.Join(output, " "))
		return
	}

	if options.System {
		output = append(output, info.System)
	}
	if options.Nodename {
		output = append(output, info.Nodename)
	}
	if options.Processor {
		output = append(output, info.Processor)
	}

	fmt.Println(strings.Join(output, " "))
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Критическая ошибка (panic): %v\n", r)
			os.Exit(1)
		}
	}()

	options, err := parseArgumentsUname()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}

	if options.Help {
		printHelpUname()
		os.Exit(0)
	}

	info, err := getUnameInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка получения информации о системе: %v\n", err)
		os.Exit(1)
	}

	printUnameInfo(info, options)

	os.Exit(0)
}
