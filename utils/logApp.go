package utils

import (
	"fmt"
	"log"
	"os"
)

var AppLogger *log.Logger

func InitLog(logPath string) {
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Errorf("ошибка открытия(создания) файла для логов: %s %w", logPath, err)
	}

	AppLogger = log.New(logFile, "APP: ", log.Ldate|log.Ltime|log.Lshortfile)
}
