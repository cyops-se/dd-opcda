package engine

import (
	"bufio"
	"compress/gzip"
	"dd-opcda/types"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
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
	createFile()
	prevRemainder = -1
	channel = make(chan types.DataMessage)
	go processMessages()
}

func CloseCache() {
	if fw != nil {
		fw.Write([]byte("]"))
		fw.Flush()
		gzw.Close()
		file.Close()
		log.Printf("CACHE: closed\n")
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
	count := 0
	proxy := proxies[2]
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

func resendCacheItem(item CacheItem, proxy *types.DiodeProxy) int {
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

			// // TEST of compression capabilities for approx 500 points/s
			// for i := 0; i < 50; i++ {
			// 	for i, _ := range msg.Points {
			// 		msg.Points[i].Name = fmt.Sprintf("%d_ThisIsALongExtraTextThatIsReally_%d", time.Now().UnixNano(), time.Now().UnixNano())
			// 	}

			// 	data, _ := json.Marshal(msg)
			// 	fw.Write(data)
			// }

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
