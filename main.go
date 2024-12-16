package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	folderName         = "new_folder"
	sessionCounterFile = "session_counter.txt"
	totalCounterFile   = "total_counter.txt"
	checkInterval      = 2 * time.Second
	welcomeText        = "Здаровки"
	welcomeFileName    = "Приветствие"
)

func main() {
	logInfo("Добро пожаловать в программу мониторинга накопителя данных и работы с папками.")
	for {
		mode := selectMode()
		switch mode {
		case 1:
			logInfo("Режим 1: Мониторинг накопителя данных.")
			runFlashDriveMonitor()
		case 2:
			logInfo("Режим 2: Ручной режим работы с папками.")
			runManualFolderMode()
		case 0:
			logWarning("Выход из программы.")
			return
		default:
			logError("Неверный выбор. Попробуйте снова.")
		}
	}
}

// Функция для выбора режима работы
func selectMode() int {
	var mode int
	logInput("Выберите режим работы:\n1 - Автоматический режим работы\n2 - Мануальный режим работы с папками\n0 - Выход")
	fmt.Scan(&mode)
	return mode
}

// Мониторинг флешек
func runFlashDriveMonitor() {
	writeSessionCounter(0)
	flashDrivePath := setDirectoryForcheck()

	for {
		if waitForFlashDrive(flashDrivePath) {
			handleFlashDrive(flashDrivePath)
			waitForFlashDriveRemoval(flashDrivePath)
		}
	}
}

func waitForFlashDrive(path string) bool {
	for {
		if canAccessFlashDrive(path) {
			return true
		}
		time.Sleep(checkInterval)
	}
}

func waitForFlashDriveRemoval(path string) {
	for {
		if !canAccessFlashDrive(path) {
			logInfo("Накопитель данных данных отключен. Ожидание нового подключения...")
			return
		}
		time.Sleep(checkInterval)
	}
}

func handleFlashDrive2(path string) {
	logAction(fmt.Sprintf("Обнаружен накопитель данных: %s", path))

	sessionCounter := readSessionCounter() + 1
	writeSessionCounter(sessionCounter)
	poemsForFun(sessionCounter)

	totalCounter := readTotalCounter() + 1
	writeTotalCounter(totalCounter)

	folderPath := filepath.Join(path, folderName)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		logInfo("Нажмите 1 для создания папки")
		openExplorer(path)
		var a string
		fmt.Scan(&a)
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil {
			logError(fmt.Sprintf("Ошибка при создании папки: %v", err))
			return
		}
		logAction("Успешное создание папки")
		createWelcomeFile(folderPath)

	}

	logInfo("Нажмите 2 для удаления папки")
	var b string
	fmt.Scan(&b)
	err := os.RemoveAll(folderPath)
	if err != nil {
		logError(fmt.Sprintf("Ошибка при удалении папки: %v", err))
		return
	}
	logAction("Папка удалена.")
	logSuccess(fmt.Sprintf("Время успешной проверки накопителя данных: %s", time.Now().Format("2006-01-02 15:04:05")))

	// err = cleanFlashDrive(path)
	// if err != nil {
	// 	logError(fmt.Sprintf("Ошибка при очистке накопителя данных: %v", err))
	// 	return
	// }
	// logSuccess("Накопитель данных очищен.")

	logInfo(fmt.Sprintf("Количество подключений в текущей сессии: %d", sessionCounter))
	logInfo(fmt.Sprintf("Общее количество подключений флешек: %d\n", totalCounter))
}

func handleFlashDrive(path string) {
	logAction(fmt.Sprintf("Обнаружен Накопитель данных: %s", path))

	folderPath := filepath.Join(path, folderName)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil {
			logError(fmt.Sprintf("Ошибка при создании папки: %v", err))
			return
		}
		logAction("Папка создана.")
	}

	createWelcomeFile(folderPath)

	err := os.RemoveAll(folderPath)
	if err != nil {
		logError(fmt.Sprintf("Ошибка при удалении папки: %v", err))
		return
	}
	logAction("Папка удалена.")
	logSuccess(fmt.Sprintf("Время успешной проверки накопителя данных: %s", time.Now().Format("2006-01-02 15:04:05")))

	// err = cleanFlashDrive(path)
	// if err != nil {
	// 	logError(fmt.Sprintf("Ошибка при очистке накопителя данных: %v", err))
	// 	return
	// }
	logSuccess("Накопитель данных очищен.")
	sessionCounter := readSessionCounter() + 1
	writeSessionCounter(sessionCounter)
	poemsForFun(sessionCounter)

	totalCounter := readTotalCounter() + 1
	writeTotalCounter(totalCounter)

	logInfo(fmt.Sprintf("Количество подключений в текущей сессии: %d", sessionCounter))
	logInfo(fmt.Sprintf("Общее количество подключений накопителя данных: %d\n", totalCounter))
}

func canAccessFlashDrive(path string) bool {
	_, err := ioutil.ReadDir(path)
	return err == nil
}

//Функция полно очистки всего содержимого на накопителе данных( временно не работает)
//func cleanFlashDrive(path string) error {
// 	files, err := ioutil.ReadDir(path)
// 	if err != nil {
// 		return err
// 	}F

// 	for _, file := range files {
// 		fullPath := filepath.Join(path, file.Name())
// 		if err := os.RemoveAll(fullPath); err != nil {
// 			logError(fmt.Sprintf("Не удалось удалить %s: %v", fullPath, err))
// 		}
// 	}

// 	return nil
// }

func setDirectoryForcheck() string {
	var disc string
	logInput("Вставьте Накопитель данных в ридер\nВведите название тома диска для проверки (например, \"H\", \"G\") или нажмите 1 для диска по умолчанию: \"H\"")
	fmt.Scan(&disc)
	if disc == "1" {
		return "H:"
	}
	logInput("Вставьте Накопитель данных")
	return disc + ":"
}

// Ручной режим работы с папками
func runManualFolderMode() {
	writeSessionCounter(0)
	flashDrivePath := setDirectoryForcheck()

	for {
		if waitForFlashDrive(flashDrivePath) {
			handleFlashDrive2(flashDrivePath)
			waitForFlashDriveRemoval(flashDrivePath)
		}
	}
}

func openExplorer(path string) {
	err := exec.Command("explorer", path).Start()
	if err != nil {
		logError(fmt.Sprintf("Ошибка при открытии проводника: %v", err))
	}
}

// Работа со счётчиками
func readSessionCounter() int {
	return readCounter(sessionCounterFile)
}

func writeSessionCounter(counter int) {
	writeCounter(sessionCounterFile, counter)
}

func readTotalCounter() int {
	return readCounter(totalCounterFile)
}

func writeTotalCounter(counter int) {
	writeCounter(totalCounterFile, counter)
}

func readCounter(filename string) int {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		logError(fmt.Sprintf("Ошибка чтения %s: %v", filename, err))
		return 0
	}
	counter, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		logError(fmt.Sprintf("Ошибка преобразования %s: %v", filename, err))
		return 0
	}
	return counter
}

func writeCounter(filename string, counter int) {
	err := ioutil.WriteFile(filename, []byte(strconv.Itoa(counter)), 0644)
	if err != nil {
		logError(fmt.Sprintf("Ошибка записи %s: %v", filename, err))
	}
}

func createWelcomeFile(folderPath string) {
	filePath := filepath.Join(folderPath, welcomeFileName)
	file, err := os.Create(filePath)
	if err != nil {
		logError(fmt.Sprintf("Ошибка при создании приветственного файла: %v", err))
		return
	}
	defer file.Close()

	_, err = file.WriteString(welcomeText)
	if err != nil {
		logError(fmt.Sprintf("Ошибка записи в файл: %v", err))
		return
	}
	logAction(fmt.Sprintf("Приветственный файл '%s' создан.", welcomeFileName))
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

// Логирование
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
