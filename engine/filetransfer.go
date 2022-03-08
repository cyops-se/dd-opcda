package engine

import (
	"crypto/sha256"
	"dd-opcda/logger"
	"dd-opcda/types"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"time"
)

type header struct {
	directory string
	filename  string
	size      int
	hashvalue []byte
}

type context struct {
	basedir       string
	newdir        string
	processingdir string
	donedir       string
	modulus       int
	msdelay       int
}

var proxy *types.DiodeProxy

func InitFileTransfer(gctx types.Context) error {
	m, _ := strconv.Atoi(InitSetting("filetransfer.modulus", "20", "Number of packets to send before quick pause").Value)
	d, _ := strconv.Atoi(InitSetting("filetransfer.msdelay", "20", "Number of milliseconds to pause before start sending again").Value)

	proxy = FirstProxy()
	if proxy != nil {
		ctx := initContext(m, d, gctx)
		go monitorFilesystem(ctx)
		return nil
	}

	return logger.Error("file transfer disabled", "No proxy defined. At least one proxy must be defined")
}

func initContext(m int, d int, gctx types.Context) *context {
	ctx := &context{basedir: path.Join(gctx.Wdir, "outgoing"), newdir: "new", processingdir: "processing", donedir: "done", modulus: m, msdelay: d}
	os.MkdirAll(path.Join(ctx.basedir, ctx.newdir), 0755)
	os.MkdirAll(path.Join(ctx.basedir, ctx.processingdir), 0755)
	os.MkdirAll(path.Join(ctx.basedir, ctx.donedir), 0755)
	return ctx
}

func monitorFilesystem(ctx *context) {
	ticker := time.NewTicker(500 * time.Millisecond)

	for {
		<-ticker.C
		processDirectory(ctx, ".")
	}
}

func processDirectory(ctx *context, dirname string) {
	readdir := path.Join(ctx.basedir, ctx.newdir, dirname)
	processingdir := path.Join(ctx.basedir, ctx.processingdir, dirname)
	os.MkdirAll(processingdir, 0755)

	infos, _ := ioutil.ReadDir(readdir)
	for _, fi := range infos {
		if !fi.IsDir() {
			filename := path.Join(readdir, fi.Name())
			movename := path.Join(processingdir, fi.Name())
			if err := os.Rename(filename, movename); err == nil {
				// log.Printf("Requested processing of file: %s (%s)", filename, movename)
				info := &types.FileInfo{Name: fi.Name(), Path: dirname, Size: int(fi.Size()), Date: fi.ModTime()}
				logger.NotifySubscribers("filetransfer.request", info)
				sendFile(ctx, info) // Do it sequentially to minimize packet loss
			} else {
				// log.Printf("Failed to move file to processing area: %s, error %s", filename, err.Error())
			}
		} else {
			processDirectory(ctx, path.Join(dirname, fi.Name()))
		}
	}

}

func sendFile(ctx *context, info *types.FileInfo) error {
	dir := info.Path
	name := info.Name
	filename := path.Join(ctx.basedir, ctx.processingdir, dir, name)

	fi, err := os.Lstat(filename)
	if err != nil {
		log.Println("Cannot find file:", filename, err.Error())
		return fmt.Errorf("file not found")
	}

	if fi.IsDir() {
		log.Println("'filename' points to a directory, not a file:", filename)
		return fmt.Errorf("directory, not file")
	}

	if fi.Size() == 0 {
		log.Println("'filename' is empty:", filename)
		return fmt.Errorf("empty file")
	}

	target := fmt.Sprintf("%s:%d", proxy.EndpointIP, proxy.FilePort)
	c, err := net.Dial("udp", target)
	if err != nil {
		log.Printf("Failed to dial %s", target)
		return err
	}

	hash := calcHash(filename)
	header := fmt.Sprintf("DD-FILETRANSFER BEGIN v2 %s %s %d %x", name, dir, info.Size, hash.Sum(nil)) // :filename:directory:size:hash:

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Failed to open %s", filename)
		return err
	}

	// Always send packets of 1200 bytes, regardless
	content := make([]byte, 1200)
	n := 0
	binhead := []byte(header)
	copy(content, binhead)
	c.Write(content)

	total := 0
	counter := uint32(0)
	for err == nil {
		// each message starts with a 4 byte sequence number, then 4 bytes of size of payload, then payload
		n, err = file.Read(content[8:])
		binary.LittleEndian.PutUint32(content, counter)
		binary.LittleEndian.PutUint32(content[4:], uint32(n))
		// sn, _ := c.Write(content[:n+8])
		sn, _ := c.Write(content) // Always write full buffer
		total += sn
		counter++

		if counter%1000 == 0 {
			percent := float64(total) / float64(info.Size) * 100.0
			// log.Printf("Progress %.2f (%d / %d)", percent, total, info.Size)
			progress := &types.FileProgress{File: info, TotalSent: total, PercentDone: percent}
			logger.NotifySubscribers("filetransfer.progress", progress)
		}

		if counter%uint32(ctx.modulus) == 0 {
			time.Sleep(time.Millisecond * time.Duration(ctx.msdelay))
		}
	}

	footer := []byte("DD-FILETRANSFER END v2")
	content = make([]byte, 1200)
	binfoot := []byte(footer)
	copy(content, binfoot)
	c.Write(content)

	c.Close()
	file.Close()

	todir := path.Join(ctx.basedir, ctx.donedir, dir)
	os.MkdirAll(todir, 0755)

	movename := path.Join(todir, name)
	if err = os.Rename(filename, movename); err == nil {
		// log.Printf("Done processing file: %s (%s)", filename, movename)
		logger.Trace("File transfer complete", "File %s, size %d transferred as requested by operator", filename, info.Size)
	} else {
		logger.Error("Failed to move file", "Error when attempting to move file after file was transferred, file %s, size %d, error %s", filename, info.Size, err.Error())
		// log.Printf("Failed to move file after processing: %s", err.Error())
	}

	logger.NotifySubscribers("filetransfer.complete", info)

	time.Sleep(time.Millisecond)

	return err
}

func calcHash(filename string) hash.Hash {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return h
}
