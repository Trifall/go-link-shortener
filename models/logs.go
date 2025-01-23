package models

import (
	"go-link-shortener/database"
	"go-link-shortener/lib"
	"log"
	"time"
)

func CreateLog(logType LogType, logSource LogSource, message string, remoteAddress string) {
	db := database.GetDB()
	if db == nil {
		log.Println(lib.ERRORS.Database)
		return
	}

	newLog := &Log{
		Timestamp: time.Now(),
		Type:      LogType(logType),
		Source:    LogSource(logSource),
		Message:   message,
	}

	result := db.Create(newLog)
	if result.Error != nil {
		log.Println("Error creating log:", result.Error)
		return
	}
}
