package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	flashDrivePath     = "H:"                  // Путь к флешке
	folderName         = "new_folder"          // Имя папки для тестирования
	sessionCounterFile = "session_counter.txt" // Файл для хранения счётчика текущей сессии
	totalCounterFile   = "total_counter.txt"   // Файл для хранения общего счётчика подключений
	checkInterval      = 2 * time.Second       // Интервал проверки подключения
)

func main() {
	logInfo("Мониторинг подключений флешек начат...")

	// Обнуляем счётчик текущей сессии
	writeSessionCounter(0)

	for {
		// Ожидаем подключения флешки
		if waitForFlashDrive(flashDrivePath) {
			// Обрабатываем подключённую флешку
			handleFlashDrive(flashDrivePath)

			// Ждём, пока флешка не будет отключена
			waitForFlashDriveRemoval(flashDrivePath)
		}
	}
}

func waitForFlashDrive(path string) bool {
	// Ожидаем появления пути, указывающего на подключённую флешку
	for {
		if canAccessFlashDrive(path) {
			return true
		}
		time.Sleep(checkInterval)
	}
}

func waitForFlashDriveRemoval(path string) {
	// Ожидаем, пока доступ к флешке не исчезнет (флешка будет отключена)
	for {
		if !canAccessFlashDrive(path) {
			logInfo("Флешка отключена. Ожидание нового подключения...")
			return
		}
		time.Sleep(checkInterval)
	}
}

func handleFlashDrive(path string) {
	logAction(fmt.Sprintf("Обнаружена флешка: %s", path))

	// Чтение/обновление счётчиков
	sessionCounter := readSessionCounter()
	sessionCounter++
	writeSessionCounter(sessionCounter)

	totalCounter := readTotalCounter()
	totalCounter++
	writeTotalCounter(totalCounter)

	// Путь для создаваемой папки
	folderPath := filepath.Join(path, folderName)

	// Проверка, существует ли папка
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// Если папка не существует, создаём её
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil {
			logError(fmt.Sprintf("Ошибка при создании папки: %v", err))
			return
		}
		logAction("Папка создана на флешке.")
	}

	// Удаляем папку сразу после создания
	err := os.RemoveAll(folderPath)
	if err != nil {
		logError(fmt.Sprintf("Ошибка при удалении папки: %v", err))
		return
	}
	logAction("Папка удалена с флешки.")

	// Очищаем всю флешку
	err = cleanFlashDrive(path)
	if err != nil {
		logError(fmt.Sprintf("Ошибка при очистке флешки: %v", err))
		return
	}
	logSuccess("Флешка очищена.")

	// Информация о количестве подключений
	logInfo(fmt.Sprintf("Количество подключений в текущей сессии: %d", sessionCounter))
	logInfo(fmt.Sprintf("Общее количество подключений флешки: %d\n", totalCounter))
}

func canAccessFlashDrive(path string) bool {
	// Проверяем доступность флешки через попытку чтения её содержимого
	_, err := ioutil.ReadDir(path)
	return err == nil
}

func readSessionCounter() int {
	// Чтение счётчика текущей сессии
	data, err := ioutil.ReadFile(sessionCounterFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Если файл не существует, начинаем с нуля
			return 0
		}
		logError(fmt.Sprintf("Ошибка при чтении файла счётчика сессии: %v", err))
		return 0
	}

	// Преобразование содержимого файла в число
	counter, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		logError(fmt.Sprintf("Ошибка преобразования счётчика сессии: %v", err))
		return 0
	}

	return counter
}

func writeSessionCounter(counter int) {
	// Запись счётчика текущей сессии в файл
	err := ioutil.WriteFile(sessionCounterFile, []byte(strconv.Itoa(counter)), 0644)
	if err != nil {
		logError(fmt.Sprintf("Ошибка при записи файла счётчика сессии: %v", err))
	}
}

func readTotalCounter() int {
	// Чтение общего счётчика
	data, err := ioutil.ReadFile(totalCounterFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Если файл не существует, начинаем с нуля
			return 0
		}
		logError(fmt.Sprintf("Ошибка при чтении файла общего счётчика: %v", err))
		return 0
	}

	// Преобразование содержимого файла в число
	counter, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		logError(fmt.Sprintf("Ошибка преобразования общего счётчика: %v", err))
		return 0
	}

	return counter
}

func writeTotalCounter(counter int) {
	// Запись общего счётчика в файл
	err := ioutil.WriteFile(totalCounterFile, []byte(strconv.Itoa(counter)), 0644)
	if err != nil {
		logError(fmt.Sprintf("Ошибка при записи файла общего счётчика: %v", err))
	}
}

func cleanFlashDrive(path string) error {
	// Очищаем всю флешку, удаляя всё её содержимое
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		fullPath := filepath.Join(path, file.Name())
		if err := os.RemoveAll(fullPath); err != nil {
			logError(fmt.Sprintf("Не удалось удалить %s: %v", fullPath, err))
		}
	}

	return nil
}

func logInfo(message string) {
	color.New(color.FgWhite).Printf("[INFO] %s\n", message)
}

func logAction(message string) {
	color.New(color.FgYellow).Printf("[ACTION] %s\n", message)
}

func logSuccess(message string) {
	color.New(color.FgCyan).Printf("[SUCCESS] %s\n", message)
}

func logError(message string) {
	color.New(color.FgRed).Printf("[ERROR] %s\n", message)
}

func logWarning(message string) {
	color.New(color.FgYellow).Printf("[WARNING] %s\n", message)
}
