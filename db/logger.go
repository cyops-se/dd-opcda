package db

import (
	"fmt"
	"log"
	"time"

	"dd-opcda/types"
)

func Log(category string, title string, msg string) string {
	entry := &types.Log{Time: time.Now().UTC(), Category: category, Title: title, Description: msg}
	DB.Create(&entry)
	text := fmt.Sprintf("%s: %s, %s", category, title, msg)
	// log.Printf("%s: %s, %s", category, title, msg)
	log.Printf(text)
	return text
}

func Trace(title string, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	text := Log("trace", title, msg)
	return fmt.Errorf(text)
}

func Error(title string, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	text := Log("error", title, msg)
	return fmt.Errorf(text)
}
