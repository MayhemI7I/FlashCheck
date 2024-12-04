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
	folderName         = "new_folder"          // Имя папки для тестирования
	sessionCounterFile = "session_counter.txt" // Файл для хранения счётчика текущей сессии
	totalCounterFile   = "total_counter.txt"   // Файл для хранения общего счётчика подключений
	checkInterval      = 2 * time.Second       // Интервал проверки подключения
)

func main() {

	logInfo("Мониторинг подключений флешек начат...")

	// Обнуляем счётчик текущей сессии
	writeSessionCounter(0)

	flashDrivePath := setDirectoryForcheck()

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
	poemsForFun(sessionCounter)

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

func setDirectoryForcheck() string {
	var disc string
	logInput(fmt.Sprintf("Введите имя диска для проверки (SD карта в проводнике \"H\",\"G\"...)"))
	logInput("Для выбора диска по умолчанию нажмите Введите \"1\", диск по умолчани. : \"H\"")
	fmt.Scan(&disc)
	if disc == "1" {
		return "H:"
	}
	return disc + ":"
}

func poemsForFun(count int) {
	switch count {
	case 50:
		logWarning(`Я научилась просто, мудро жить,
Смотреть на небо и молиться Богу,
И долго перед вечером бродить,
Чтоб утомить ненужную тревогу.`)
	case 100:
		logWarning(`Шут, шут! Я встретил здесь
В лесу шута - шута в ливрее пестрой.
О, жалкий мир! Да, это верно так,
Как то, что я живу посредством пищи,
Шут встречен мной. Лежал он на земле
И, греючись на солнце, тут же
Сударыню-фортуну он честил
Хорошими, разумными словами,
А между тем он просто пестрый шут.
"Здорово, шут!" - сказал я. "Нет уж, сударь, -
Он отвечал, - не называйте вы
Меня шутом, пока богатства небо
Мне не пошлет". Затем полез в карман
И, вытащив часы, бесцветным взглядом
На них взглянул и мудро произнес:
"Десятый час! - И вслед за тем прибавил:
- Здесь видим мы, как двигается мир:
Всего лишь час назад был час девятый,
А час пройдет - одиннадцать пробьет;
И так-то вот мы с каждым часом зреем,
И так-то вот гнием мы каждый час.
И тут конец всей сказочке". Чуть только
Услышал я, что этот пестрый шут
О времени так рассуждает - печень
Моя сейчас запела петухом
От радости, что водятся такие
Мыслители среди шутов, и я
Час целый по его часам смеялся.`)
	case 200:
		logWarning(`Когда теряет равновесие
Твое сознание усталое,
Когда ступени этой лестницы
Уходят из-под ног,
Как палуба,
Когда плюет на человечество
Твое ночное одиночество, —
Ты можешь
Размышлять о вечности
И сомневаться в непорочности
Идей, гипотез, восприятия
Произведения искусства,
И — кстати — самого зачатия
Мадонной сына Иисуса.
Но лучше поклоняться данности
С ее глубокими могилами,
Которые потом,
За давностью,
Покажутся такими милыми.
Да. Лучше поклоняться данности
С короткими ее дорогами,
Которые потом
До странности
Покажутся тебе
Широкими,
Покажутся большими,
Пыльными,
Усеянными компромиссами,
Покажутся большими крыльями,
Покажутся большими птицами.
Да. Лучше поклоняться данности
С убогими ее мерилами,
Которые потом,
По крайности,
Послужат для тебя перилами,
(Хотя и не особо чистыми),
Удерживающими в равновесии
Твои хромающие истины
На этой выщербленной лестнице.`)
		return

	}
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
func logInput(message string) {
	color.New(color.FgGreen).Printf("[INPUT] %s\n", message)
}

func logWarning(message string) {
	color.New(color.FgHiGreen).Printf("[WARNING] %s\n", message)
}
