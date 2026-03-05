package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// KillFlags содержит флаги для управления kill
type KillFlags struct {
	Help   bool   // -h --help справка
	Signal string // -s    сигнал для отправки
	List   string // -l список/преобразование сигналов
	Table  bool   // -L, таблица всех сигналов
	PIDs   []string
}

// signalMap содержит карту названий сигналов на их номера
var signalMap = map[string]int{
	"HUP":      int(syscall.SIGHUP),
	"INT":      int(syscall.SIGINT),
	"QUIT":     int(syscall.SIGQUIT),
	"ILL":      int(syscall.SIGILL),
	"TRAP":     int(syscall.SIGTRAP),
	"ABRT":     int(syscall.SIGABRT),
	"BUS":      int(syscall.SIGBUS),
	"FPE":      int(syscall.SIGFPE),
	"KILL":     int(syscall.SIGKILL),
	"USR1":     int(syscall.SIGUSR1),
	"SEGV":     int(syscall.SIGSEGV),
	"USR2":     int(syscall.SIGUSR2),
	"PIPE":     int(syscall.SIGPIPE),
	"ALRM":     int(syscall.SIGALRM),
	"TERM":     int(syscall.SIGTERM),
	"STKFLT":   int(syscall.SIGSTKFLT),
	"CHLD":     int(syscall.SIGCHLD),
	"CONT":     int(syscall.SIGCONT),
	"STOP":     int(syscall.SIGSTOP),
	"TSTP":     int(syscall.SIGTSTP),
	"TTIN":     int(syscall.SIGTTIN),
	"TTOU":     int(syscall.SIGTTOU),
	"URG":      int(syscall.SIGURG),
	"XCPU":     int(syscall.SIGXCPU),
	"XFSZ":     int(syscall.SIGXFSZ),
	"VTALRM":   int(syscall.SIGVTALRM),
	"PROF":     int(syscall.SIGPROF),
	"WINCH":    int(syscall.SIGWINCH),
	"IO":       int(syscall.SIGPOLL),
	"PWR":      int(syscall.SIGPWR),
	"SYS":      int(syscall.SIGSYS),
	"RTMIN":    34,
	"RTMIN+1":  35,
	"RTMIN+2":  36,
	"RTMIN+3":  37,
	"RTMIN+4":  38,
	"RTMIN+5":  39,
	"RTMIN+6":  40,
	"RTMIN+7":  41,
	"RTMIN+8":  42,
	"RTMIN+9":  43,
	"RTMIN+10": 44,
	"RTMIN+11": 45,
	"RTMIN+12": 46,
	"RTMIN+13": 47,
	"RTMIN+14": 48,
	"RTMIN+15": 49,
	"RTMAX-14": 50,
	"RTMAX-13": 51,
	"RTMAX-12": 52,
	"RTMAX-11": 53,
	"RTMAX-10": 54,
	"RTMAX-9":  55,
	"RTMAX-8":  56,
	"RTMAX-7":  57,
	"RTMAX-6":  58,
	"RTMAX-5":  59,
	"RTMAX-4":  60,
	"RTMAX-3":  61,
	"RTMAX-2":  62,
	"RTMAX-1":  63,
	"RTMAX":    64,
}

var reverseSignalMap map[int]string

// initReverseSignalMap инициализирует обратную карту сигналов
func initReverseSignalMap() {
	if reverseSignalMap != nil {
		return
	}
	reverseSignalMap = make(map[int]string)
	for name, num := range signalMap {
		reverseSignalMap[num] = name
	}
}

// showHelp выводит справку по использованию команды kill
func showHelp() {
	fmt.Println(`Использование: kill [КЛЮЧИ] [PID ...]

Ключи:
  -h                  показать эту справку
  --help              показать эту справку
  -s СИГНАЛ           отправить сигнал
  -l [СИГНАЛ]         список сигналов или преобразование
  -L                  таблица всех сигналов

Описание:
  kill отправляет сигнал процессу. По умолчанию отправляется TERM (15).
  Сигнал можно указать по имени (KILL, TERM) или номером (9, 15).

Примеры:
  ./kill 1234               Отправить TERM сигнал PID 1234
  ./kill -9 1234            Отправить KILL сигнал PID 1234
  ./kill -s TERM 1234 5678  Отправить TERM нескольким PID
  ./kill -l                 Показать список всех сигналов
  ./kill -l 9               Преобразовать номер 9 в название
  ./kill -l TERM            Преобразовать название TERM в номер
  ./kill -L                 Показать таблицу всех сигналов`)
}

// parseFlags парсит флаги для kill
func parseFlags(args []string) (KillFlags, error) {
	flags := KillFlags{}
	var i int

	for i = 0; i < len(args); i++ {
		arg := args[i]

		// Проверяем аргумент на наличие флагов
		if strings.HasPrefix(arg, "-") && arg != "-" {
			// Обработка флагов вида -9, -KILL
			if len(arg) > 1 && arg[1] != '-' && !isLetterFlag(arg) {
				// Это может быть номер сигнала или название
				flags.Signal = arg[1:]
				continue
			}

			switch arg {
			case "-h", "--help":
				flags.Help = true
			case "-s":
				if i+1 < len(args) {
					i++
					flags.Signal = args[i]
				} else {
					return flags, fmt.Errorf("флаг %s требует аргумента", arg)
				}
			case "-l":
				// -l может быть с аргументом или без
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					flags.List = args[i]
				} else {
					flags.List = ""
				}
			case "-L":
				flags.Table = true
			default:
				return flags, fmt.Errorf("неизвестный флаг: %s", arg)
			}
		} else {
			break
		}
	}

	// Собираем все оставшиеся PID
	flags.PIDs = args[i:]

	return flags, nil
}

// isLetterFlag проверяет является ли флаг буквенным
func isLetterFlag(arg string) bool {
	if len(arg) < 2 {
		return false
	}
	// Проверяем второй символ
	ch := arg[1]
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// signalToNumber преобразует название или номер сигнала в номер
func signalToNumber(sig string) (int, error) {
	sig = strings.TrimSpace(strings.ToUpper(sig))

	// Проверяем в карте названий
	if num, ok := signalMap[sig]; ok {
		return num, nil
	}

	num, err := strconv.Atoi(sig)
	if err != nil {
		return 0, fmt.Errorf("неизвестный сигнал: %s", sig)
	}

	if num < 1 || num > 64 {
		return 0, fmt.Errorf("сигнал вне диапазона 1-64: %d", num)
	}

	return num, nil
}

// numberToSignal преобразует номер сигнала в название
func numberToSignal(num int) (string, error) {
	initReverseSignalMap()
	if name, ok := reverseSignalMap[num]; ok {
		return name, nil
	}
	return "", fmt.Errorf("неизвестный сигнал: %d", num)
}

// parsePID парсит и валидирует PID
func parsePID(pidStr string) (int, error) {
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("неверный PID '%s': %v", pidStr, err)
	}
	if pid <= 0 {
		return 0, fmt.Errorf("PID должен быть > 0: %d", pid)
	}
	return pid, nil
}

// sendSignalToPID отправляет сигнал одному процессу
func sendSignalToPID(pid int, signum int) error {
	err := syscall.Kill(pid, syscall.Signal(signum))
	if err != nil {
		return fmt.Errorf("PID %d: %v", pid, err)
	}
	return nil
}

// sendSignalToPIDs отправляет сигнал списку процессов
func sendSignalToPIDs(pids []int, signum int) error {
	var errors []string

	for _, pid := range pids {
		if err := sendSignalToPID(pid, signum); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

// listSignals выводит список сигналов или преобразует сигнал
func listSignals(arg string) error {
	initReverseSignalMap()

	if arg == "" {
		// Выводим список всех сигналов в формате как в оригинале
		var nums []int
		for num := range reverseSignalMap {
			nums = append(nums, num)
		}

		// Сортируем по номерам
		n := len(nums)
		for i := 0; i < n-1; i++ {
			for j := 0; j < n-i-1; j++ {
				if nums[j] > nums[j+1] {
					nums[j], nums[j+1] = nums[j+1], nums[j]
				}
			}
		}

		// Выводим в 5 колонок как в оригинале
		for i, num := range nums {
			sigName := reverseSignalMap[num]
			fmt.Printf("%2d) SIG%-8s", num, sigName)
			if (i+1)%5 == 0 {
				fmt.Println()
			} else {
				fmt.Print("\t")
			}
		}
		fmt.Println()
		return nil
	}

	// Пробуем как номер и выводим название
	sigNum, err := strconv.Atoi(arg)
	if err == nil {
		if name, ok := reverseSignalMap[sigNum]; ok {
			fmt.Println(name)
			return nil
		}
	}

	// Пробуем преобразовать как название сигнала в номер
	num, err := signalToNumber(arg)
	if err == nil {
		fmt.Println(num)
		return nil
	}

	return fmt.Errorf("неизвестный сигнал: %s", arg)
}

// printSignalTable выводит таблицу всех сигналов
func printSignalTable() {
	initReverseSignalMap()

	fmt.Printf("%-6s %-10s\n", "Номер", "Название")
	fmt.Println("------------------")

	// Собираем и сортируем номера сигналов
	var nums []int
	for num := range reverseSignalMap {
		nums = append(nums, num)
	}

	n := len(nums)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if nums[j] > nums[j+1] {
				nums[j], nums[j+1] = nums[j+1], nums[j]
			}
		}
	}

	for _, num := range nums {
		fmt.Printf("%-6d %-10s\n", num, reverseSignalMap[num])
	}
}

// executeKill выполняет основную логику команды kill
func executeKill(flags KillFlags) error {
	// Парсим все PID
	var pids []int
	for _, pidStr := range flags.PIDs {
		pid, err := parsePID(pidStr)
		if err != nil {
			return err
		}
		pids = append(pids, pid)
	}

	// Определяем сигнал
	signum := int(syscall.SIGTERM)
	if flags.Signal != "" {
		var err error
		signum, err = signalToNumber(flags.Signal)
		if err != nil {
			return err
		}
	}

	return sendSignalToPIDs(pids, signum)
}

func main() {
	flags, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "kill: %v\n", err)
		os.Exit(1)
	}

	// Приоритет обработки опций
	if flags.Help {
		showHelp()
		return
	}

	if flags.Table {
		printSignalTable()
		return
	}

	// Проверяем -l ПЕРЕД проверкой наличия PID
	if flags.List != "" || (len(os.Args) > 1 && os.Args[1] == "-l") {
		listArg := flags.List
		if len(os.Args) > 2 {
			listArg = os.Args[2]
		}
		if err := listSignals(listArg); err != nil {
			fmt.Fprintf(os.Stderr, "kill: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Проверяем наличие PID для отправки сигнала
	if len(flags.PIDs) == 0 {
		fmt.Fprintf(os.Stderr, "kill: требуется хотя бы 1 PID\n")
		os.Exit(1)
	}

	// Выполняем kill
	if err := executeKill(flags); err != nil {
		fmt.Fprintf(os.Stderr, "kill: %v\n", err)
		os.Exit(1)
	}
}
