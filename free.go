package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// FreeFlags содержит флаги для управления free
type FreeFlags struct {
	Help      bool // --help справка
	Human     bool // -h человекочитаемый формат
	Kilobytes bool // -k размеры в кибибайтах
	Megabytes bool // -m размеры в мебибайтах
	Gigabytes bool // -g размеры в гибибайтах
}

// MemInfo содержит информацию о памяти из /proc/meminfo
type MemInfo struct {
	MemTotal     uint64 // всего оперативной памяти
	MemFree      uint64 // свободная память
	MemAvailable uint64 // доступная память для новых процессов
	Buffers      uint64 // память под буферы
	Cached       uint64 // кэшированная память
	SwapTotal    uint64 // всего swap памяти
	SwapFree     uint64 // свободно swap памяти
	SReclaimable uint64 // освобождаемая память slab
}

// MemoryStats содержит статистику по памяти
type MemoryStats struct {
	Total     uint64 // всего
	Used      uint64 // использовано
	Free      uint64 // свободно
	Shared    uint64 // общее
	BuffCache uint64 // буферы и кэш
	Available uint64 // доступно
}

// showHelp выводит справку по использованию команды free
func showHelp() {
	fmt.Println(`Использование: free [КЛЮЧИ]

Ключи:
  -h              выводить размеры в человекочитаемом формате
  -k              выводить размеры в кибибайтах
  -m              выводить размеры в мебибайтах (MB)
  -g              выводить размеры в гибибайтах (GB)
  --help          показать эту справку

Описание:
  free отображает информацию об используемой и свободной оперативной памяти
  и swap-пространстве в системе.

Примеры:
  ./free -h       Показать с автоматическим выбором единиц
  ./free -k       Показать память в кибибайтах
  ./free -m       Показать в мебибайтах
  ./free -g       Показать в гибибайтах`)
}

// parseFlags парсит флаги для free
func parseFlags(args []string) (FreeFlags, error) {
	flags := FreeFlags{
		Kilobytes: true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") && arg != "-" {
			switch arg {
			case "--help":
				flags.Help = true
			case "-h":
				flags.Human = true
				flags.Kilobytes = false
				flags.Megabytes = false
				flags.Gigabytes = false
			case "-k":
				flags.Kilobytes = true
				flags.Human = false
				flags.Megabytes = false
				flags.Gigabytes = false
			case "-m":
				flags.Megabytes = true
				flags.Kilobytes = false
				flags.Human = false
				flags.Gigabytes = false
			case "-g":
				flags.Gigabytes = true
				flags.Kilobytes = false
				flags.Human = false
				flags.Megabytes = false
			default:
				return flags, fmt.Errorf("неизвестный флаг: %s", arg)
			}
		} else {
			return flags, fmt.Errorf("неожиданный аргумент: %s", arg)
		}
	}

	return flags, nil
}

// readMemInfo читает информацию о памяти из /proc/meminfo
func readMemInfo() (*MemInfo, error) {
	memInfo := &MemInfo{}
	memData := make(map[string]uint64)

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть /proc/meminfo: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Ищем двоеточие разделяющее имя и значение
		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			continue
		}

		// Извлекаем имя параметра
		key := strings.TrimSpace(line[:colonIndex])

		// Извлекаем значение и единицу измерения
		valueStr := strings.TrimSpace(line[colonIndex+1:])
		parts := strings.Fields(valueStr)

		if len(parts) == 0 {
			continue
		}

		value, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			continue
		}

		// Конвертируем килобайты в байты
		if len(parts) == 2 && parts[1] == "kB" {
			value = value * 1024
		}

		memData[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения /proc/meminfo: %v", err)
	}

	memInfo.MemTotal = memData["MemTotal"]
	memInfo.MemFree = memData["MemFree"]
	memInfo.MemAvailable = memData["MemAvailable"]
	memInfo.Buffers = memData["Buffers"]
	memInfo.Cached = memData["Cached"]
	memInfo.SwapTotal = memData["SwapTotal"]
	memInfo.SwapFree = memData["SwapFree"]
	memInfo.SReclaimable = memData["SReclaimable"]

	if memInfo.MemTotal == 0 {
		return nil, fmt.Errorf("некорректные данные в /proc/meminfo")
	}

	return memInfo, nil
}

// calculateMemoryStats вычисляет статистику памяти
func calculateMemoryStats(memInfo *MemInfo) (*MemoryStats, *MemoryStats) {
	// Статистика по основной памяти
	ramStats := &MemoryStats{
		Total:     memInfo.MemTotal,
		Free:      memInfo.MemFree,
		BuffCache: memInfo.Buffers + memInfo.Cached + memInfo.SReclaimable,
		Available: memInfo.MemAvailable,
		Shared:    0,
	}

	ramStats.Used = ramStats.Total - ramStats.Free - ramStats.BuffCache

	// Статистика по swap-памяти
	swapStats := &MemoryStats{
		Total: memInfo.SwapTotal,
		Free:  memInfo.SwapFree,
	}

	swapStats.Used = swapStats.Total - swapStats.Free

	return ramStats, swapStats
}

// formatHumanReadable форматирует размер в человекочитаемом формате
func formatHumanReadable(size uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.1fTi", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.1fGi", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1fMi", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1fKi", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%dB", size)
	}
}

// formatSize форматирует размер согласно флагам
func formatSize(size uint64, flags FreeFlags) string {
	switch {
	case flags.Human:
		return formatHumanReadable(size)
	case flags.Megabytes:
		sizeMB := size / (1024 * 1024)
		return fmt.Sprintf("%d", sizeMB)
	case flags.Gigabytes:
		sizeGB := size / (1024 * 1024 * 1024)
		return fmt.Sprintf("%d", sizeGB)
	default:
		sizeKB := size / 1024
		return fmt.Sprintf("%d", sizeKB)
	}
}

// getUnit возвращает единицу измерения для вывода
func getUnit(flags FreeFlags) string {
	switch {
	case flags.Human:
		return ""
	case flags.Megabytes:
		return "MB"
	case flags.Gigabytes:
		return "GB"
	default:
		return "KiB"
	}
}

// displayMemoryInfo выводит информацию о памяти в таблице
func displayMemoryInfo(ramStats, swapStats *MemoryStats, flags FreeFlags) {
	unit := getUnit(flags)

	// Заголовок таблицы
	fmt.Printf("              total        used        free      shared  buff/cache   available\n")

	// Строка для основной памяти
	fmt.Printf("Mem:    %12s %11s %11s %11s %11s %11s\n",
		formatSize(ramStats.Total, flags),
		formatSize(ramStats.Used, flags),
		formatSize(ramStats.Free, flags),
		formatSize(ramStats.Shared, flags),
		formatSize(ramStats.BuffCache, flags),
		formatSize(ramStats.Available, flags))

	// Строка для swap памяти
	fmt.Printf("Swap:   %12s %11s %11s\n",
		formatSize(swapStats.Total, flags),
		formatSize(swapStats.Used, flags),
		formatSize(swapStats.Free, flags))

	// Выводим единицу измерения если это не человекочитаемый формат
	if unit != "" {
		fmt.Printf("\nЕдиницы измерения: %s\n", unit)
	}
}

// executeFree выполняет основную логику команды free
func executeFree(flags FreeFlags) error {
	// Читаем информацию о памяти
	memInfo, err := readMemInfo()
	if err != nil {
		return err
	}

	// Вычисляем статистику
	ramStats, swapStats := calculateMemoryStats(memInfo)

	// Выводим информацию
	displayMemoryInfo(ramStats, swapStats, flags)

	return nil
}

func main() {
	flags, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "free: %v\n", err)
		os.Exit(1)
	}

	if flags.Help {
		showHelp()
		return
	}

	if err := executeFree(flags); err != nil {
		fmt.Fprintf(os.Stderr, "free: %v\n", err)
		os.Exit(1)
	}
}
