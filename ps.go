package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Структура для хранения параметров команды
type PsOptions struct {
	help         bool
	allProcesses bool  // -A, -e все процессы
	allExceptBg  bool  // -a все кроме фоновых
	filterPIDs   []int // -p фильтр по PID
}

// Структура для хранения информации о процессе
type ProcessInfo struct {
	PID   int    // ID процесса
	Comm  string // Имя команды
	TTY   int    // Терминал
	UID   int    // ID пользователя
	UTime uint64 // User time
	STime uint64 // System time
}

type ProcessMap map[int]*ProcessInfo

// Slice для хранения отфильтрованных процессов
type ProcessList []*ProcessInfo

// Реализуем интерфейс sort.Interface для сортировки по PID
func (p ProcessList) Len() int           { return len(p) }
func (p ProcessList) Less(i, j int) bool { return p[i].PID < p[j].PID }
func (p ProcessList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Функция для вывода справки
func printHelpPs() {
	helpText := `Использование: ps [КЛЮЧИ]
  
Ключи:
  -h, --help    Показать справку
  -A, -e        Выбрать все процессы в системе
  -a            Выбрать все процессы, кроме фоновых
  -p PID...     Выбрать процессы по PID

Описание:
  ps - показывает информацию о запущенных процессах в системе.
            
Примеры:
  ps -A           Показать все процессы
  ps -a           Показать процессы с терминалом
  ps -p 1         Показать информацию о процессе с PID 1
  ps -p 1 2 3     Показать информацию о нескольких процессах`
	fmt.Println(helpText)
}

// Функция для чтения списка всех PID из /proc
func readAllPIDs() ([]int, error) {
	var pids []int

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения /proc: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		pids = append(pids, pid)
	}

	if len(pids) == 0 {
		return nil, fmt.Errorf("не найдено ни одного процесса в /proc")
	}

	return pids, nil
}

// Функция для чтения информации о процессе из /proc/[pid]/stat
func readProcessStat(pid int) (*ProcessInfo, error) {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)

	data, err := os.ReadFile(statPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения %s: %v", statPath, err)
	}

	statStr := string(data)

	// Находим позиции скобок для извлечения comm
	commStart := strings.Index(statStr, "(")
	commEnd := strings.LastIndex(statStr, ")")

	if commStart == -1 || commEnd == -1 || commEnd <= commStart {
		return nil, fmt.Errorf("неверный формат файла stat")
	}

	// Извлекаем части строки
	pidStr := strings.TrimSpace(statStr[:commStart])
	comm := statStr[commStart+1 : commEnd]
	restStr := strings.TrimSpace(statStr[commEnd+1:])

	// Разбиваем оставшуюся часть на поля
	fields := strings.Fields(restStr)

	// Проверяем минимальное количество полей
	if len(fields) < 15 {
		return nil, fmt.Errorf("недостаточно полей в stat файле")
	}

	procInfo := &ProcessInfo{
		Comm: comm,
	}

	procInfo.PID, err = strconv.Atoi(pidStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга PID: %v", err)
	}

	procInfo.TTY, _ = strconv.Atoi(fields[4])
	procInfo.UTime, _ = strconv.ParseUint(fields[11], 10, 64)
	procInfo.STime, _ = strconv.ParseUint(fields[12], 10, 64)
	procInfo.UID, _ = readProcessUID(pid)

	return procInfo, nil
}

// Функция для чтения UID процесса из /proc/[pid]/status
func readProcessUID(pid int) (int, error) {
	statusPath := fmt.Sprintf("/proc/%d/status", pid)

	data, err := os.ReadFile(statusPath)
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения %s: %v", statusPath, err)
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				uid, err := strconv.Atoi(fields[1])
				if err != nil {
					return 0, err
				}
				return uid, nil
			}
		}
	}

	return 0, fmt.Errorf("UID не найден в status файле")
}

// Функция для чтения информации о всех процессах
func readAllProcesses() (ProcessMap, error) {
	processes := make(ProcessMap)

	pids, err := readAllPIDs()
	if err != nil {
		return nil, err
	}

	successCount := 0

	for _, pid := range pids {
		procInfo, err := readProcessStat(pid)
		if err != nil {
			continue
		}

		processes[pid] = procInfo
		successCount++
	}

	if successCount == 0 {
		return nil, fmt.Errorf("не удалось прочитать информацию ни об одном процессе")
	}

	return processes, nil
}

// Получаем TTY текущего процесса
func getCurrentTTY() int {
	data, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		return 0
	}

	statStr := string(data)
	commEnd := strings.LastIndex(statStr, ")")
	if commEnd == -1 {
		return 0
	}

	restStr := strings.TrimSpace(statStr[commEnd+1:])
	fields := strings.Fields(restStr)
	if len(fields) < 5 {
		return 0
	}

	tty, _ := strconv.Atoi(fields[4])
	return tty
}

// Функция для фильтрации процессов согласно опциям
func filterProcesses(processes ProcessMap, options *PsOptions) (ProcessList, error) {
	var filtered ProcessList

	// Если указан фильтр по PID
	if len(options.filterPIDs) > 0 {
		for _, pid := range options.filterPIDs {
			if proc, exists := processes[pid]; exists {
				filtered = append(filtered, proc)
			} else {
				fmt.Fprintf(os.Stderr, "Предупреждение: процесс с PID %d не найден\n", pid)
			}
		}

		sort.Sort(filtered)
		return filtered, nil
	}

	if options.allProcesses {
		for _, proc := range processes {
			filtered = append(filtered, proc)
		}
		sort.Sort(filtered)
		return filtered, nil
	}

	if options.allExceptBg {
		for _, proc := range processes {
			if proc.TTY != 0 {
				filtered = append(filtered, proc)
			}
		}
		sort.Sort(filtered)
		return filtered, nil
	}

	// По умолчанию - показываем процессы текущего пользователя на текущем терминале
	currentUID := os.Getuid()
	currentTTY := getCurrentTTY()

	for _, proc := range processes {
		if proc.UID == currentUID && proc.TTY == currentTTY && currentTTY != 0 {
			filtered = append(filtered, proc)
		}
	}

	sort.Sort(filtered)
	return filtered, nil
}

// Функция для форматирования времени в формат HH:MM:SS
func formatTime(utime, stime uint64) string {
	totalTicks := utime + stime
	seconds := totalTicks / 100

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

// Функция для преобразования tty_nr в имя терминала
func formatTTY(ttyNr int) string {
	if ttyNr == 0 {
		return "?"
	}

	major := (ttyNr >> 8) & 0xff
	minor := (ttyNr & 0xff) | ((ttyNr >> 12) & 0xfff00)

	switch major {
	case 136: // pts
		return fmt.Sprintf("pts/%d", minor)
	case 4: // tty
		if minor < 64 {
			return fmt.Sprintf("tty%d", minor)
		}
		return fmt.Sprintf("ttyS%d", minor-64)
	case 5: // console/ptmx
		if minor == 0 {
			return "tty"
		}
		if minor == 1 {
			return "console"
		}
		if minor == 2 {
			return "ptmx"
		}
	}

	return fmt.Sprintf("%d/%d", major, minor)
}

// Функция для вывода информации о процессах
func printProcesses(processes ProcessList) {
	if len(processes) == 0 {
		fmt.Println("Процессы не найдены")
		return
	}

	// Выводим заголовок таблицы в формате как оригинальный ps
	fmt.Printf("%7s %-8s %8s %s\n", "PID", "TTY", "TIME", "CMD")

	// Выводим информацию о каждом процессе
	for _, proc := range processes {
		ttyStr := formatTTY(proc.TTY)
		timeStr := formatTime(proc.UTime, proc.STime)
		fmt.Printf("%7d %-8s %8s %s\n",
			proc.PID,
			ttyStr,
			timeStr,
			proc.Comm,
		)
	}
}

// Функция для валидации и парсинга аргументов
func parseArgumentsPs() (*PsOptions, error) {
	options := &PsOptions{}

	flagSet := flag.NewFlagSet("ps", flag.ContinueOnError)
	flagSet.SetOutput(os.Stdout)

	var pidStr string
	flagSet.BoolVar(&options.help, "h", false, "Показать справку")
	flagSet.BoolVar(&options.help, "help", false, "Показать справку")
	flagSet.BoolVar(&options.allProcesses, "A", false, "Выбрать все процессы")
	flagSet.BoolVar(&options.allProcesses, "e", false, "Выбрать все процессы (синоним -A)")
	flagSet.BoolVar(&options.allExceptBg, "a", false, "Выбрать все процессы кроме фоновых")
	flagSet.StringVar(&pidStr, "p", "", "Выбрать процессы по PID")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("ошибка разбора аргументов: %v", err)
	}

	// Парсим PID если указан флаг -p
	if pidStr != "" {
		pidsToProcess := []string{pidStr}

		// Добавляем оставшиеся аргументы если это числа
		for _, arg := range flagSet.Args() {
			if _, err := strconv.Atoi(arg); err == nil {
				pidsToProcess = append(pidsToProcess, arg)
			} else {
				// Если встретили нечисло, это ошибка
				return nil, fmt.Errorf("неожиданный аргумент: %s", arg)
			}
		}

		// Парсим все PID
		for _, pidArg := range pidsToProcess {
			// Если в строке есть запятые разбиваем по запятым
			if strings.Contains(pidArg, ",") {
				parts := strings.Split(pidArg, ",")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if part != "" {
						pid, err := strconv.Atoi(part)
						if err != nil {
							return nil, fmt.Errorf("неверный формат PID: %s", part)
						}
						if pid <= 0 {
							return nil, fmt.Errorf("PID должен быть положительным числом: %d", pid)
						}
						options.filterPIDs = append(options.filterPIDs, pid)
					}
				}
			} else {
				pid, err := strconv.Atoi(pidArg)
				if err != nil {
					return nil, fmt.Errorf("неверный формат PID: %s", pidArg)
				}
				if pid <= 0 {
					return nil, fmt.Errorf("PID должен быть положительным числом: %d", pid)
				}
				options.filterPIDs = append(options.filterPIDs, pid)
			}
		}
	} else if flagSet.NArg() > 0 {
		return nil, fmt.Errorf("неожиданные аргументы: %v\nИспользуйте -h для справки", flagSet.Args())
	}

	// Проверяем взаимоисключающие опции
	optionsCount := 0
	if options.allProcesses {
		optionsCount++
	}
	if options.allExceptBg {
		optionsCount++
	}
	if len(options.filterPIDs) > 0 {
		optionsCount++
	}

	if optionsCount > 1 {
		return nil, fmt.Errorf("нельзя использовать несколько опций фильтрации одновременно")
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

	options, err := parseArgumentsPs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}

	if options.help {
		printHelpPs()
		os.Exit(0)
	}

	processes, err := readAllProcesses()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения процессов: %v\n", err)
		os.Exit(1)
	}

	filtered, err := filterProcesses(processes, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка фильтрации: %v\n", err)
		os.Exit(1)
	}

	printProcesses(filtered)

	os.Exit(0)
}
