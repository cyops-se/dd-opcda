package main

import (
	"dd-opcda/db"
	"dd-opcda/engine"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
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
	target     string
	port       int
	source     string
	progID     string
	config     string
	create     bool
	list       bool
	branch     string
	iterations int
	interval   int
	cmd        string
	trace      bool
}

var ctx Context

func main() {
	defer handlePanic()

	svcName := "dd-opcda"
	flag.StringVar(&ctx.target, "target", "172.26.8.243", "The address of the outer inserter")
	flag.IntVar(&ctx.port, "port", 4357, "The UDP port of the outer inserter")
	flag.StringVar(&ctx.source, "source", "localhost", "The address of the OPC server")
	flag.StringVar(&ctx.progID, "progid", "IntegrationObjects.AdvancedSimulator.1", "The OPC server prog id")
	flag.StringVar(&ctx.config, "c", "", "Configuration file, for example declaring tags to process. If not specified, all tags will be processed")
	flag.BoolVar(&ctx.create, "create", false, "Use this parameter to create a config file with all tags found in the specified server")
	flag.BoolVar(&ctx.list, "list", false, "Lists the OPC DA servers available on the specified source")
	flag.StringVar(&ctx.branch, "branch", "", "Lists all tags at the specified branch tag")
	flag.IntVar(&ctx.iterations, "i", 1, "Number of times to get all specified tags (used to measure performance)")
	flag.IntVar(&ctx.interval, "p", 1, "Read interval in seconds")
	flag.StringVar(&ctx.cmd, "cmd", "debug", "Windows service command (try 'usage' for more info)")
	flag.BoolVar(&ctx.trace, "trace", false, "Prints traces of OCP data to the console")
	flag.Parse()

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

	// Sleep until interrupted
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Exiting (waiting 1 sec) ...")
	time.Sleep(time.Second * 1)
}

func runEngine() {
	defer handlePanic()

	db.ConnectDatabase()
	engine.InitGroups()
	engine.InitServers()
	go RunWeb()
}
