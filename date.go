package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// DateFlags содержит флаги для управления date
type DateFlags struct {
	Help bool   // -h/--help справка
	UTC  bool   // -u использовать UTC
	Date string // -d строка даты
	File string // -f файл с датами
}

// showHelp выводит справку по использованию команды date
func showHelp() {
	fmt.Println(`Использование: date [КЛЮЧИ]

Ключи:
  -d СТРОКА       показать время, описанное СТРОКОЙ вместо текущего
  -f ФАЙЛ         как -d для каждой строки из ФАЙЛА
  -u              использовать UTC время
  -h              показать эту справку
  --help          показать эту справку

Описание:
  date - выводит текущую дату и время.

Примеры:
  ./date                        Показать текущую дату и время
  ./date -u                     Показать UTC время
  ./date -d "2006-01-02"        Показать дату из строки
  ./date -f date.txt           Показать даты из файла`)
}

// parseFlags парсит флаги для date
func parseFlags(args []string) (DateFlags, error) {
	flags := DateFlags{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") && arg != "-" {
			if arg == "--help" || arg == "-h" {
				flags.Help = true
			} else if arg == "-u" {
				flags.UTC = true
			} else if arg == "-d" {
				if i+1 < len(args) {
					i++
					flags.Date = args[i]
				} else {
					return flags, fmt.Errorf("флаг -d требует аргумента")
				}
			} else if arg == "-f" {
				if i+1 < len(args) {
					i++
					flags.File = args[i]
				} else {
					return flags, fmt.Errorf("флаг -f требует аргумента")
				}
			} else if len(arg) > 2 {
				for _, ch := range arg[1:] {
					switch ch {
					case 'u':
						flags.UTC = true
					case 'h':
						flags.Help = true
					default:
						return flags, fmt.Errorf("неизвестный флаг: -%c", ch)
					}
				}
			} else {
				return flags, fmt.Errorf("неизвестный флаг: %s", arg)
			}
		}
	}

	return flags, nil
}

// parseDate парсит строку даты
func parseDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"15:04:05",
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"02 Jan 2006",
		"02 Jan 2006 15:04:05",
		time.RFC3339,
		time.RFC822,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			if layout == "15:04:05" {
				now := time.Now()
				t = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
			} else if layout == "2006-01-02" {
				t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("не удалось разобрать дату: %s", dateStr)
}

// formatTime форматирует время date
func formatTime(t time.Time) string {
	// Массивы дней недели и месяцев
	days := []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}
	months := []string{"янв", "фев", "мар", "апр", "май", "июн", "июл", "авг", "сен", "окт", "ноя", "дек"}

	day := days[t.Weekday()]
	month := months[t.Month()-1]
	date := t.Day()
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()
	tz := t.Format("MST")
	year := t.Year()

	return fmt.Sprintf("%s %2d %s %d %02d:%02d:%02d %s", day, date, month, year, hour, minute, second, tz)
}

func main() {
	flags, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "date: %v\n", err)
		fmt.Fprintln(os.Stderr, "Используй date --help для справки")
		os.Exit(1)
	}

	if flags.Help {
		showHelp()
		return
	}

	// Обработка файла с датами
	if flags.File != "" {
		file, err := os.Open(flags.File)
		if err != nil {
			fmt.Fprintf(os.Stderr, "date: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			t, err := parseDate(line)
			if err != nil {
				fmt.Fprintf(os.Stderr, "date: %v\n", err)
				continue
			}

			if flags.UTC {
				t = t.UTC()
			}

			fmt.Println(formatTime(t))
		}
		return
	}

	// Определяем время
	var t time.Time
	if flags.Date != "" {
		var err error
		t, err = parseDate(flags.Date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "date: %v\n", err)
			os.Exit(1)
		}
	} else {
		t = time.Now()
	}

	// Конвертируем в UTC если нужно
	if flags.UTC {
		t = t.UTC()
	}

	// Выводим время
	fmt.Println(formatTime(t))
}
