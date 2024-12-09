package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
)

const (
	folderName         = "new_folder"
	sessionCounterFile = "session_counter.txt"
	totalCounterFile   = "total_counter.txt"
	checkInterval      = 2 * time.Second
)

var (
	exitProgram bool
	exitMutex   sync.Mutex
)

func main() {
	initializeLogger()
	logInfo("Добро пожаловать в программу мониторинга флешек и работы с папками.")

	go listenForExit() // Горутин для отслеживания клавиши Esc

	currentMode := 1 // По умолчанию режим мониторинга
	for {
		if shouldExit() {
			logWarning("Выход из программы.")
			return
		}

		switch currentMode {
		case 1:
			logInfo("Режим 1: Мониторинг флешек.")
			currentMode = runFlashDriveMonitorWithSwitch()
		case 2:
			logInfo("Режим 2: Ручной режим работы с папками.")
			currentMode = runManualFolderModeWithSwitch()
		default:
			logError("Неверный выбор режима.")
			currentMode = 1
		}
	}
}

// Функция для отслеживания клавиши Esc
func listenForExit() {
	if err := keyboard.Open(); err != nil {
		log.Fatalf("Ошибка при открытии клавиатуры: %v", err)
	}
	defer keyboard.Close()

	for {
		_, key, err := keyboard.GetKey()
		if err != nil {
			logError(fmt.Sprintf("Ошибка чтения клавиши: %v", err))
			continue
		}

		if key == keyboard.KeyEsc {
			setExitFlag(true)
			break
		}
	}
}

// Проверка флага выхода
func shouldExit() bool {
	exitMutex.Lock()
	defer exitMutex.Unlock()
	return exitProgram
}

// Установка флага выхода
func setExitFlag(flag bool) {
	exitMutex.Lock()
	defer exitMutex.Unlock()
	exitProgram = flag
}

// Неблокирующий выбор режима
func selectModeNonBlocking() int {
	logInfo("Выберите новый режим (1 или 2). Для продолжения текущего нажмите Enter.")
	timeout := time.After(5 * time.Second)
	inputChan := make(chan string)

	go func() {
		var input string
		fmt.Scanln(&input)
		inputChan <- input
	}()

	select {
	case input := <-inputChan:
		if input == "1" || input == "2" {
			mode, _ := strconv.Atoi(input)
			return mode
		}
	case <-timeout:
		logInfo("Время выбора режима истекло. Продолжаем текущий режим.")
	}

	return 1 // Возврат к мониторингу по умолчанию
}

// Мониторинг флешек с возможностью смены режима
func runFlashDriveMonitorWithSwitch() int {
	flashDrivePath := setDirectoryForCheck()
	for {
		if shouldExit() {
			return 0 // Завершение программы
		}

		if waitForFlashDrive(flashDrivePath) {
			handleFlashDrive(flashDrivePath)
			waitForFlashDriveRemoval(flashDrivePath)

			// Проверка на смену режима без блокировки
			return selectModeNonBlocking()
		}
	}
}

// Ручной режим работы с возможностью смены режима
func runManualFolderModeWithSwitch() int {
	flashDrivePath := setDirectoryForCheck()
	for {
		if shouldExit() {
			return 0 // Завершение программы
		}

		if waitForFlashDrive(flashDrivePath) {
			handleFlashDriveInteractive(flashDrivePath)
			waitForFlashDriveRemoval(flashDrivePath)

			// Проверка на смену режима без блокировки
			return selectModeNonBlocking()
		}
	}
}

// Ожидание подключения флешки
func waitForFlashDrive(path string) bool {
	for {
		if shouldExit() {
			return false
		}

		if canAccessFlashDrive(path) {
			return true
		}
		time.Sleep(checkInterval)
	}
}

// Ожидание отключения флешки
func waitForFlashDriveRemoval(path string) {
	for {
		if shouldExit() {
			return
		}

		if !canAccessFlashDrive(path) {
			logInfo("SD карта отключена. Ожидание нового подключения...")
			return
		}
		time.Sleep(checkInterval)
	}
}

// Основная обработка флешки
func handleFlashDrive(path string) {
	handleFlashDriveGeneric(path, false)
}

// Обработка флешки в интерактивном режиме
func handleFlashDriveInteractive(path string) {
	handleFlashDriveGeneric(path, true)
}

// Универсальная обработка флешки
func handleFlashDriveGeneric(path string, interactive bool) {
	logAction(fmt.Sprintf("Обнаружена SD карта: %s", path))

	folderPath := filepath.Join(path, folderName)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil {
			logError(fmt.Sprintf("Ошибка при создании папки: %v", err))
			return
		}
		logAction("Папка создана на флешке.")
	}

	if interactive {
		logInfo("Нажмите Enter для удаления папки.")
		fmt.Scanln()
	}

	err := os.RemoveAll(folderPath)
	if err != nil {
		logError(fmt.Sprintf("Ошибка при удалении папки: %v", err))
		return
	}
	logAction("Папка удалена с SD карты.")
}

// Проверка доступа к флешке
func canAccessFlashDrive(path string) bool {
	_, err := ioutil.ReadDir(path)
	return err == nil
}

// Настройка директории для проверки
func setDirectoryForCheck() string {
	for {
		logInput("Введите имя диска для проверки (например, H, G) или нажмите 1 для диска по умолчанию: H")
		var disc string
		fmt.Scan(&disc)

		if disc == "1" {
			disc = "H"
		}
		path := disc + ":"
		if canAccessFlashDrive(path) {
			logSuccess(fmt.Sprintf("Выбрана директория: %s", path))
			return path
		}
		logError(fmt.Sprintf("Диск %s недоступен. Попробуйте снова.", path))
	}
}

// Логирование
func initializeLogger() {
	logFile, err := os.OpenFile("program.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Не удалось создать лог-файл: %v", err)
	}
	log.SetOutput(logFile)
}

func logInfo(message string) {
	color.New(color.FgWhite).Printf("[INFO] %s\n", message)
	log.Println("[INFO] " + message)
}

func logAction(message string) {
	color.New(color.FgYellow).Printf("[ACTION] %s\n", message)
	log.Println("[ACTION] " + message)
}

func logSuccess(message string) {
	color.New(color.FgCyan).Printf("[SUCCESS] %s\n", message)
	log.Println("[SUCCESS] " + message)
}

func logError(message string) {
	color.New(color.FgRed).Printf("[ERROR] %s\n", message)
	log.Println("[ERROR] " + message)
}

func logInput(message string) {
	color.New(color.FgGreen).Printf("[INPUT] %s\n", message)
}

func logWarning(message string) {
	color.New(color.FgHiGreen).Printf("[WARNING] %s\n", message)
	log.Println("[WARNING] " + message)
}
