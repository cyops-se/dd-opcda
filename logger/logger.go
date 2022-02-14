package logger

import (
	"fmt"
	"log"
	"time"

	"dd-opcda/db"
	"dd-opcda/types"
)

func Log(category string, title string, msg string) string {
	entry := &types.Log{Time: time.Now().UTC(), Category: category, Title: title, Description: msg}
	db.DB.Create(&entry)
	NotifySubscribers(fmt.Sprintf("logger.%s", category), entry)
	text := fmt.Sprintf("%s: %s, %s", category, title, msg)
	log.Printf(text)
	return text
}

func Trace(title string, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	NotifySubscribers("logger.trace", msg)
	text := Log("trace", title, msg)
	return fmt.Errorf(text)
}

func Error(title string, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	text := Log("error", title, msg)
	return fmt.Errorf(text)
}
