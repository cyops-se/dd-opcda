package engine

import (
	"crypto/sha256"
	"dd-opcda/db"
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

func InitFileTransfer() error {
	proxy = FirstProxy()
	if proxy != nil {
		ctx := initContext()
		log.Println("SENDER")
		go monitorFilesystem(ctx)
		return nil
	}

	return db.Error("file transfer disabled", "No proxy defined. At least one proxy must be defined")
}

func initContext() *context {
	ctx := &context{basedir: "outgoing", newdir: "new", processingdir: "processing", donedir: "done", modulus: 20, msdelay: 20}
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
				log.Printf("Requested processing of file: %s (%s)", filename, movename)
				info := &types.FileInfo{Name: fi.Name(), Path: dirname, Size: int(fi.Size()), Date: fi.ModTime()}
				NotifySubscribers("filetransfer.request", info)
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

	target := fmt.Sprintf("%s:%d", proxy.EndpointIP, proxy.FilePort)
	c, err := net.Dial("udp", target)
	if err != nil {
		log.Printf("Failed to dial %s", target)
		return err
	}

	hash := calcHash(filename)
	header := fmt.Sprintf("DD-FILETRANSFER  %s %s %d %x", name, dir, info.Size, hash.Sum(nil)) // :filename:directory:size:hash:

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Failed to open %s", filename)
		return err
	}

	content := make([]byte, 1400)
	n := 0

	// Send header in single packet
	c.Write([]byte(header))
	time.Sleep(time.Millisecond)

	total := 0
	counter := uint32(0)
	for err == nil {
		binary.LittleEndian.PutUint32(content, counter)
		n, err = file.Read(content[4:])
		sn, _ := c.Write(content[:n+4])
		total += sn
		counter++

		if counter%1000 == 0 {
			percent := float64(total) / float64(info.Size) * 100.0
			log.Printf("Progress %.2f (%d / %d)", percent, total, info.Size)
			progress := &types.FileProgress{File: info, TotalSent: total, PercentDone: percent}
			NotifySubscribers("filetransfer.progress", progress)
		}

		if counter%uint32(ctx.modulus) == 0 {
			time.Sleep(time.Millisecond * time.Duration(ctx.msdelay))
		}
	}

	c.Close()
	file.Close()

	todir := path.Join(ctx.basedir, ctx.donedir, dir)
	os.MkdirAll(todir, 0755)

	movename := path.Join(todir, name)
	if err = os.Rename(filename, movename); err == nil {
		log.Printf("Done processing file: %s (%s)", filename, movename)
	} else {
		log.Printf("Failed to move file after processing: %s", err.Error())
	}

	NotifySubscribers("filetransfer.complete", info)

	time.Sleep(time.Millisecond)

	return nil
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
