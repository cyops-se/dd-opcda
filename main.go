package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/konimarti/opc"
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
	trace bool
}

var ctx Context

func main() {
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

	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if !isIntSess {
		runService(svcName, false)
		return
	}

	switch ctx.cmd {
	case "debug":
		runEngine()
		return
	case "install":
		err = installService(svcName, svcName)
	case "remove":
		err = removeService(svcName)
	case "start":
		err = startService(svcName)
	case "stop":
		err = controlService(svcName, svc.Stop, svc.StopPending)
	case "pause":
		err = controlService(svcName, svc.Pause, svc.Paused)
	case "continue":
		err = controlService(svcName, svc.Continue, svc.Running)
	default:
		usage(fmt.Sprintf("invalid command %s", ctx.cmd))
	}
	if err != nil {
		log.Fatalf("failed to %s %s: %v", ctx.cmd, svcName, err)
	}
}

func runEngine() {
	log.Println("Running OPC engine")

	if ctx.list {
		if ao := opc.NewAutomationObject(); ao != nil {
			servers_found := ao.GetOPCServers(ctx.source)
			log.Printf("Found %d server(s) on '%s':\n", len(servers_found), ctx.source)
			for _, server := range servers_found {
				log.Println(server)
			}
		} else {
			log.Println("Unable to get new automation object")
		}
		return
	}

	if ctx.create && ctx.config == "" {
		ctx.config = "config.json"
	}

	
	tags := []string{""}
	if ctx.config == "" {
		browser, err := opc.CreateBrowser(ctx.progID, // ProgId
			[]string{ctx.source}, // Nodes
		)

		if err != nil {
			log.Println("Failed to browse OPC server, error:", err)
			return
		}

		log.Println("Available tags")
		if ctx.branch != "" {
			subtree := opc.ExtractBranchByName(browser, ctx.branch)
			if subtree == nil {
				log.Println("No tags available at specified branch:", ctx.branch)
				return
			}

			opc.PrettyPrint(subtree)
			tags = opc.CollectTags(subtree)
		} else {
			opc.PrettyPrint(browser)
			tags = opc.CollectTags(browser)
		}
	}

	if ctx.create {
		c := &Config{}
		c.Tags = make([]ConfigTagEntry, len(tags))
		for t, tag := range tags {
			c.Tags[t].Name = tag
		}
		if data, err := json.MarshalIndent(c, "", "    "); err == nil {
			if err = ioutil.WriteFile(ctx.config, data, 0644); err != nil {
				log.Printf("Unable to write JSON config file, %s. Error: %v\n", ctx.config, err)
			} else {
				log.Println("New configuration file with all tags created:", ctx.config)
				return
			}
		} else {
			log.Printf("Unable to create JSON for config file, %s. Error: %v\n", ctx.config, err)
		}
		return
	}

	// Check config and use tags defined there if any
	if ctx.config != "" {
		if data, err := ioutil.ReadFile(ctx.config); err == nil {
			c := &Config{}
			if err = json.Unmarshal(data, &c); err == nil && len(c.Tags) > 0 {
				tags = make([]string, len(c.Tags))
				for i, tag := range c.Tags {
					tags[i] = tag.Name
				}
			} else {
				log.Printf("Unable to interpret config file, %s. Will be using all tags found in the server. Error: %v\n", ctx.config, err)
			}
		} else {
			log.Printf("Unable to open specified config file, %s. Will be using all tags found in the server. Error: %v\n", ctx.config, err)
		}
	}

	if len(tags) <= 0 {
		log.Println("No tags defined, aborting...")
		return
	}

	client, err := opc.NewConnection(ctx.progID, // ProgId
		[]string{ctx.source}, //  OPC servers nodes
		tags,                 // slice of OPC tags
	)
	defer client.Close()

	if err != nil {
		log.Println("Failed to connect to OPC server, error:", err)
		return
	}

	con, err := net.Dial("udp", fmt.Sprintf("%s:%d", ctx.target, ctx.port))
	if con == nil {
		log.Println("Failed to connect target server:", err)
		return
	}

	defer con.Close()

	log.Println("Connected...")
	timer := time.NewTicker(time.Duration(ctx.interval) * time.Second)

	// read all added tags
	go func() {
		items := client.Read() // This is only to get the number of items
		msg := &DataMessage{}
		msg.Count = 10
		msg.Points = make([]DataPoint, msg.Count)

		for {
			<-timer.C
			start := float64(time.Now().UnixNano())
			for j := 0; j < ctx.iterations; j++ {
				items = client.Read()

				if ctx.trace {
					log.Println("items:", items)
				}

				var i, b int // golang always initialize to 0
				for k, v := range items {
					msg.Points[b].Time = v.Timestamp
					msg.Points[b].Name = k
					msg.Points[b].Value = v.Value
					msg.Points[b].Quality = int(v.Quality)

					// Send batch when msg.Points is full (keep it small to avoid fragmentation)
					if b == len(msg.Points)-1 || i == len(items)-1 {
						data, _ := json.Marshal(msg)
						con.Write(data)
						b = 0
						msg.Counter++
					} else {
						b++
					}
					i++
				}
			}
			end := float64(time.Now().UnixNano())
			if ctx.iterations > 1 {
				log.Printf("It took %f seconds to read %d tags\n", (end-start)/1000000000.0, len(items)*ctx.iterations)
			}
		}
	}()

	// Sleep until interrupted
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Exiting ...")
	client.Close()
}
