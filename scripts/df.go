package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
)

// DFFlags содержит флаги для управления df
type DFFlags struct {
	Help          bool // --help справка
	HumanReadable bool // -h размеры в понятном формате степени 1024
	SI            bool // -H размеры в понятном формате степени 1000
	All           bool // -a показать все ФС, включая виртуальные
	Type          bool // -T выводить тип файловой системы
	Kilobytes     bool // -k размеры в килобайтах
}

// FileSystemInfo содержит информацию о файловой системе
type FileSystemInfo struct {
	Filesystem  string // название устройства
	MountPoint  string // точка монтирования
	FSType      string // тип ФС
	TotalSize   uint64 // всего байт
	UsedSize    uint64 // использовано байт
	AvailSize   uint64 // доступно байт
	UsedPercent int    // процент использования
}

// MountPoint содержит информацию о точке монтирования из /proc/mounts
type MountPoint struct {
	Device     string
	MountPoint string
	FSType     string
	Options    string
}

// DeviceID идентифицирует уникальное устройство по номерам
type DeviceID struct {
	Major uint64
	Minor uint64
}

// showHelp выводит справку по использованию команды df
func showHelp() {
	fmt.Println(`Использование: df [КЛЮЧИ]

Ключи:
  -h                    выводить размеры в понятном формате (K, M, G, T) степени 1024
  -H                    выводить размеры в понятном формате (K, M, G, T) степени 1000
  -a                    показать все файловые системы, включая виртуальные
  -T                    выводить тип файловой системы
  -k                    выводить размеры в килобайтах
  --help                показать эту справку

Описание:
  df отображает информацию об использовании дискового пространства
  для смонтированных файловых систем.

Примеры:
  ./df                    Показать основные файловые системы в килобайтах
  ./df -h                 Показать размеры в формате K, M, G (1024)
  ./df -H                 Показать размеры в формате K, M, G (1000)
  ./df -a                 Показать все файловые системы
  ./df -T                 Показать тип файловой системы
  ./df -hT                Показать размеры и типы ФС`)
}

// parseFlags парсит флаги для df
func parseFlags(args []string) (DFFlags, error) {
	flags := DFFlags{
		Kilobytes: true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") && arg != "-" {
			if arg == "--help" {
				flags.Help = true
			} else if strings.HasPrefix(arg, "--") {
				return flags, fmt.Errorf("неизвестный флаг: %s", arg)
			} else {
				for _, ch := range arg[1:] {
					switch ch {
					case 'h':
						flags.HumanReadable = true
						flags.SI = false
						flags.Kilobytes = false
					case 'H':
						flags.SI = true
						flags.HumanReadable = false
						flags.Kilobytes = false
					case 'a':
						flags.All = true
					case 'T':
						flags.Type = true
					case 'k':
						flags.Kilobytes = true
						flags.HumanReadable = false
						flags.SI = false
					default:
						return flags, fmt.Errorf("неизвестный флаг: -%c", ch)
					}
				}
			}
		} else {
			return flags, fmt.Errorf("неожиданный аргумент: %s", arg)
		}
	}

	return flags, nil
}

// readMountPoints читает список смонтированных файловых систем
func readMountPoints() ([]MountPoint, error) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть /proc/mounts: %v", err)
	}
	defer file.Close()

	var mounts []MountPoint
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		if len(fields) < 4 {
			continue
		}

		mounts = append(mounts, MountPoint{
			Device:     fields[0],
			MountPoint: fields[1],
			FSType:     fields[2],
			Options:    fields[3],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения /proc/mounts: %v", err)
	}

	return mounts, nil
}

// shouldIncludeFS проверяет нужно ли включать эту ФС в вывод
func shouldIncludeFS(fsType, device string, showAll bool) bool {
	if showAll {
		return true
	}

	// Виртуальные ФС которые не показываются без -a
	dummyFS := map[string]bool{
		"sysfs":       true,
		"proc":        true,
		"devpts":      true,
		"cgroup":      true,
		"cgroup2":     true,
		"pstore":      true,
		"bpf":         true,
		"configfs":    true,
		"debugfs":     true,
		"tracefs":     true,
		"securityfs":  true,
		"selinuxfs":   true,
		"fusectl":     true,
		"mqueue":      true,
		"hugetlbfs":   true,
		"autofs":      true,
		"rpc_pipefs":  true,
		"nfsd":        true,
		"binfmt_misc": true,
		"nsfs":        true,
		"ramfs":       true,
		"efivarfs":    true,
		"rootfs":      true,
		"sockfs":      true,
		"pipefs":      true,
	}

	if dummyFS[fsType] {
		return false
	}

	// Разрешаем tmpfs и devtmpfs
	if fsType == "tmpfs" || fsType == "devtmpfs" {
		return true
	}

	// Разрешаем только реальные устройства из dev
	return strings.HasPrefix(device, "/dev/")
}

// getFileSystemInfo получает информацию о файловой системе
func getFileSystemInfo(mount MountPoint) (*FileSystemInfo, error) {
	var stat syscall.Statfs_t

	if err := syscall.Statfs(mount.MountPoint, &stat); err != nil {
		return nil, fmt.Errorf("ошибка statfs для %s: %v", mount.MountPoint, err)
	}

	blockSize := uint64(stat.Bsize)
	if stat.Bsize < 0 {
		blockSize = uint64(-stat.Bsize)
	}

	totalSize := stat.Blocks * blockSize
	availSize := stat.Bavail * blockSize
	freeSize := stat.Bfree * blockSize
	usedSize := totalSize - freeSize

	var usedPercent int
	usedPlusAvail := usedSize + availSize
	if usedPlusAvail > 0 {
		usedPercent = int((usedSize*100 + usedPlusAvail - 1) / usedPlusAvail)
	}

	return &FileSystemInfo{
		Filesystem:  mount.Device,
		MountPoint:  mount.MountPoint,
		FSType:      mount.FSType,
		TotalSize:   totalSize,
		UsedSize:    usedSize,
		AvailSize:   availSize,
		UsedPercent: usedPercent,
	}, nil
}

// getDeviceID получает уникальный идентификатор устройства для отслеживания дубликатов
func getDeviceID(mountPoint string) (DeviceID, error) {
	var stat syscall.Stat_t

	if err := syscall.Stat(mountPoint, &stat); err != nil {
		return DeviceID{}, fmt.Errorf("ошибка stat для %s: %v", mountPoint, err)
	}

	dev := stat.Dev
	return DeviceID{
		Major: uint64(dev >> 8),
		Minor: uint64(dev & 0xFF),
	}, nil
}

// formatSize форматирует размер согласно флагам
func formatSize(size uint64, flags DFFlags) string {
	switch {
	case flags.HumanReadable:
		return formatHumanReadable1024(size)
	case flags.SI:
		return formatHumanReadable1000(size)
	default:
		sizeKB := size / 1024
		return fmt.Sprintf("%d", sizeKB)
	}
}

// formatHumanReadable1024 форматирует размер в понятном формате степени 1024
func formatHumanReadable1024(size uint64) string {
	units := []string{"", "K", "M", "G", "T"}
	divisor := 1024.0
	sizeFloat := float64(size)

	for i := 0; i < len(units); i++ {
		if sizeFloat < divisor || i == len(units)-1 {
			if i == 0 {
				return fmt.Sprintf("%.0f", sizeFloat)
			}
			if sizeFloat == float64(int64(sizeFloat)) {
				return fmt.Sprintf("%.0f%s", sizeFloat, units[i])
			}
			return fmt.Sprintf("%.1f%s", sizeFloat, units[i])
		}
		sizeFloat /= divisor
	}

	return fmt.Sprintf("%.0f", sizeFloat)
}

// formatHumanReadable1000 форматирует размер в понятном формате степени 1000
func formatHumanReadable1000(size uint64) string {
	units := []string{"", "K", "M", "G", "T"}
	divisor := 1000.0
	sizeFloat := float64(size)

	for i := 0; i < len(units); i++ {
		if sizeFloat < divisor || i == len(units)-1 {
			if i == 0 {
				return fmt.Sprintf("%.0f", sizeFloat)
			}
			// Для степени 1000 используем одну десятичную если число не целое
			if sizeFloat == float64(int64(sizeFloat)) {
				return fmt.Sprintf("%.0f%s", sizeFloat, units[i])
			}
			return fmt.Sprintf("%.1f%s", sizeFloat, units[i])
		}
		sizeFloat /= divisor
	}

	return fmt.Sprintf("%.0f", sizeFloat)
}

// getUnit возвращает единицу измерения для заголовка таблицы
func getUnit(flags DFFlags) string {
	switch {
	case flags.HumanReadable:
		return "Размер"
	case flags.SI:
		return "Размер"
	default:
		return "1K-блоки"
	}
}

// truncate обрезает строку до заданной длины с многоточием
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// displayFileSystemInfo выводит информацию о файловых системах в таблице
func displayFileSystemInfo(infos []FileSystemInfo, flags DFFlags) {
	unit := getUnit(flags)

	if flags.Type {
		// Заголовок таблицы с типом ФС
		fmt.Printf("%-20s %-10s %12s %12s %12s %5s %s\n",
			"Файловая система", "Тип", unit, "Использовано", "Доступно", "Исп%", "Смонтировано в")

		// Строки таблицы
		for _, info := range infos {
			fmt.Printf("%-20s %-10s %12s %12s %12s %4d%% %s\n",
				truncate(info.Filesystem, 20),
				info.FSType,
				formatSize(info.TotalSize, flags),
				formatSize(info.UsedSize, flags),
				formatSize(info.AvailSize, flags),
				info.UsedPercent,
				info.MountPoint)
		}
	} else {
		// Заголовок таблицы без типа ФС
		fmt.Printf("%-20s %12s %12s %12s %5s %s\n",
			"Файловая система", unit, "Использовано", "Доступно", "Исп%", "Смонтировано в")

		// Строки таблицы
		for _, info := range infos {
			fmt.Printf("%-20s %12s %12s %12s %4d%% %s\n",
				truncate(info.Filesystem, 20),
				formatSize(info.TotalSize, flags),
				formatSize(info.UsedSize, flags),
				formatSize(info.AvailSize, flags),
				info.UsedPercent,
				info.MountPoint)
		}
	}
}

// executeDf выполняет основную логику команды df
func executeDf(flags DFFlags) error {
	mounts, err := readMountPoints()
	if err != nil {
		return err
	}

	var infos []FileSystemInfo
	seenDevices := make(map[DeviceID]bool)

	// Обрабатываем каждую точку монтирования
	for _, mount := range mounts {
		// Проверяем нужно ли включать эту ФС
		if !shouldIncludeFS(mount.FSType, mount.Device, flags.All) {
			continue
		}

		// Получаем device ID для отслеживания дубликатов
		devID, err := getDeviceID(mount.MountPoint)
		if err != nil {
			if flags.All {
				fmt.Fprintf(os.Stderr, "Предупреждение: %v\n", err)
			}
			continue
		}

		// Пропускаем если уже видели это устройство
		if seenDevices[devID] {
			continue
		}
		seenDevices[devID] = true

		// Получаем информацию о ФС
		info, err := getFileSystemInfo(mount)
		if err != nil {
			if flags.All {
				fmt.Fprintf(os.Stderr, "Предупреждение: %v\n", err)
			}
			continue
		}

		infos = append(infos, *info)
	}

	if len(infos) == 0 {
		return fmt.Errorf("не найдено доступных файловых систем")
	}

	displayFileSystemInfo(infos, flags)
	return nil
}

func main() {
	flags, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "df: %v\n", err)
		os.Exit(1)
	}

	if flags.Help {
		showHelp()
		return
	}

	if err := executeDf(flags); err != nil {
		fmt.Fprintf(os.Stderr, "df: %v\n", err)
		os.Exit(1)
	}
}
