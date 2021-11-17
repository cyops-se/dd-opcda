package main

import (
	"dd-opcda/db"
	"dd-opcda/engine"
	"dd-opcda/routes"
	"flag"
	"fmt"
	"log"
	"time"

	"golang.org/x/sys/windows/svc"
)

type DataPoint struct {
	Time    time.Time   `json:"t"`
	Name    string      `json:"n"`
	Value   interface{} `json:"v"`
	Quality int         `json:"q"`
}

type DataMessage struct {
	Counter uint64      `json:"counter"`
	Count   int         `json:"count"`
	Points  []DataPoint `json:"points"`
}

type ConfigTagEntry struct {
	Name string `json:"name"`
}

type Config struct {
	Tags []ConfigTagEntry `json:"tags"`
}

type Context struct {
	cmd     string
	trace   bool
	version bool
}

var ctx Context
var GitVersion string
var GitCommit string

func main() {
	defer handlePanic()

	svcName := "dd-opcda"
	flag.StringVar(&ctx.cmd, "cmd", "debug", "Windows service command (try 'usage' for more info)")
	flag.BoolVar(&ctx.trace, "trace", false, "Prints traces of OCP data to the console")
	flag.BoolVar(&ctx.version, "v", false, "Prints the commit hash and exists")
	flag.Parse()

	routes.SysInfo.GitVersion = GitVersion
	routes.SysInfo.GitCommit = GitCommit

	if ctx.version {
		fmt.Printf("dd-opcda version %s, commit: %s\n", routes.SysInfo.GitVersion, routes.SysInfo.GitCommit)
		return
	}

	if ctx.cmd == "install" {
		if err := installService(svcName, "dd-opcda from cyops-se"); err != nil {
			log.Fatalf("failed to %s %s: %v", ctx.cmd, svcName, err)
		}
		return
	} else if ctx.cmd == "remove" {
		if err := removeService(svcName); err != nil {
			log.Fatalf("failed to %s %s: %v", ctx.cmd, svcName, err)
		}
		return
	}

	inService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if we are running in service: %v", err)
	}
	if inService {
		runService(svcName, false)
		return
	}

	// runEngine()
	runService(svcName, true)
	engine.CloseCache()
}

func runEngine() {
	defer handlePanic()

	db.ConnectDatabase()
	engine.InitGroups()
	engine.InitServers()
	engine.InitCache()
	engine.InitFileTransfer()
	go RunWeb()
}
