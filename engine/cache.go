package engine

import (
	"bufio"
	"compress/gzip"
	"dd-opcda/logger"
	"dd-opcda/types"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var channel chan types.DataMessage
var file *os.File
var gzw *gzip.Writer
var fw *bufio.Writer
var firstWrite bool
var prevRemainder int

type CacheItem struct {
	Filename string    `json:"filename"`
	Time     time.Time `json:"time"`
	Size     int64     `json:"size"`
}

type CacheInfo struct {
	Size      int64       `json:"size"`
	Count     int         `json:"count"`
	FirstTime time.Time   `json:"firsttime"`
	LastTime  time.Time   `json:"lasttime"`
	Items     []CacheItem `json:"items"`
}

// var cacheItems []CacheItem
// var cacheSize int64
var cacheInfo CacheInfo
var cacheMutex sync.Mutex

func InitCache() {
	InitSetting("cache.retention", "7", "Number of days to retain cached files")

	createFile()
	prevRemainder = -1
	channel = make(chan types.DataMessage)
	go processMessages()
	go pruneCache()
}

func CloseCache() {
	if fw != nil {
		fw.Write([]byte("]"))
		fw.Flush()
		gzw.Close()
		file.Close()
	}
}

func GetCacheInfo() CacheInfo {
	refreshCache()
	if cacheInfo.Count > 0 {
		cacheInfo.FirstTime = cacheInfo.Items[0].Time
		cacheInfo.LastTime = cacheInfo.Items[cacheInfo.Count-1].Time
	}

	return cacheInfo
}

func ResendCacheItems(items []CacheItem) int {
	if cacheInfo.Size <= 0 {
		// Refresh the cache info
		GetCacheInfo()
	}

	count := 0
	if proxy == nil {
		log.Printf("No proxy defined")
		return 0
	}

	for _, item := range items {
		for _, fi := range cacheInfo.Items {
			if fi.Filename == item.Filename {
				count += resendCacheItem(item, proxy)
				break
			}
		}
	}

	return count
}

func SendFullCache() error {
	// Just copy the files to the outgoing file transfer directory
	return copy("cache", "outgoing/new")
}

func copy(source, destination string) error {
	var err error = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		var relPath string = strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(destination, relPath), 0755)
		} else {
			var data, err1 = ioutil.ReadFile(filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return ioutil.WriteFile(filepath.Join(destination, relPath), data, 0777)
		}
	})
	if err != nil {
		log.Println("copy command failed: ", err.Error())
	}
	return err
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	dir := path.Dir(strings.ReplaceAll(dst, "\\", "/"))
	log.Printf("Creating directory %s", dir)
	os.MkdirAll(dir, 0755)
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func resendCacheItem(item CacheItem, proxy *types.DiodeProxy) int {
	// First put it on the file transfer directory
	newFilename := path.Join("outgoing", "new", item.Filename)
	if err := copyFile(item.Filename, newFilename); err != nil {
		log.Printf("Failed to copy file %s to %s, error: %s", item.Filename, newFilename, err.Error())
	}
	/*
		// Secondly, read it and send each message again
		file, _ = os.OpenFile(item.Filename, os.O_RDONLY, 0644)
		gzr, _ := gzip.NewReader(file)
		fr := bufio.NewReader(gzr)

		count := 0
		data, _ := ioutil.ReadAll(fr)
		var msgs []types.DataMessage
		if err := json.Unmarshal(data, &msgs); err == nil {
			for _, msg := range msgs {
				data, _ := json.Marshal(msg)
				proxy.DataChan <- data
				count++
				time.Sleep(time.Millisecond)
			}
		} else {
			log.Println("Failed to unmarshal to JSON, error:", err.Error())
		}

		gzr.Close()
		file.Close()

		return count
	*/
	return 1
}

func getTimeFromFilename(filename string) time.Time {
	var year, day, hour, minute int
	var month time.Month
	fmt.Sscanf(filename, "dd_%d_%02d_%02d-%02d_%02d.json.gz", &year, &month, &day, &hour, &minute)
	t := time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
	return t
}

func indexer(p string, info os.FileInfo, err error) error {
	if !info.IsDir() {
		item := &CacheItem{Filename: p, Time: getTimeFromFilename(info.Name()), Size: info.Size()}
		cacheInfo.Items = append(cacheInfo.Items, *item)
		cacheInfo.Size += info.Size()
	}
	return nil
}

func refreshCache() {
	cacheMutex.Lock()
	cacheInfo.Items = nil
	cacheInfo.Size = 0
	if err := filepath.Walk("cache", indexer); err != nil {
		log.Println("FILEWALK ERROR:", err.Error())
	}
	cacheInfo.Count = len(cacheInfo.Items)
	cacheMutex.Unlock()
}

func cacheMessage(msg *types.DataMessage) {
	channel <- *msg
}

func processMessages() {
	for {
		msg := <-channel
		remainder := time.Now().UTC().Minute() % 5 // New file every 5 minutes
		if remainder == 0 && remainder != prevRemainder {
			createFile()
		}

		prevRemainder = remainder

		if fw != nil {
			if firstWrite {
				fw.Write([]byte("["))
			} else {
				fw.Write([]byte(","))
			}

			data, _ := json.Marshal(msg)
			fw.Write(data)

			firstWrite = false
		}
	}
}

func createFile() {
	now := time.Now().UTC()
	dirpath := fmt.Sprintf("cache/%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	filename := fmt.Sprintf("dd_%d_%02d_%02d-%02d_%02d.json.gz", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
	fullname := path.Join(dirpath, filename)

	os.MkdirAll(dirpath, os.ModePerm)

	CloseCache()

	// If the file doesn't exist, create it, or append to the file
	file, _ = os.OpenFile(fullname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	gzw = gzip.NewWriter(file)
	fw = bufio.NewWriter(gzw)
	firstWrite = true
}

func pruneCache() {
	days := 7

	// Check once an hour
	ticker := time.NewTicker(time.Second * 10)
	for {
		<-ticker.C

		if s, err := GetSetting("cache.retention"); err == nil {
			days, _ = strconv.Atoi(s.Value)
			if days < 1 {
				days = 7
			}
		}

		utc := time.Now().UTC()
		refreshCache()
		count := 0
		for _, item := range cacheInfo.Items {
			if utc.Sub(item.Time) > time.Duration(uint64(days)*24*uint64(time.Hour)) {
				os.Remove(item.Filename)
				count++
			}
		}

		if count > 0 {
			logger.Trace("Cache pruned", "%d files pruned from cache", count)
		}
	}
}
