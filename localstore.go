package main

import "os"

type LocalStore struct {
	basePath    string
	currentPath string
	fileName    string
	file        *os.File
}
