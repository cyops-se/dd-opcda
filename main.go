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

func main() {
	target := flag.String("target", "172.16.11.3", "The address of the outer inserter")
	port := flag.Int("port", 4357, "The UDP port of the outer inserter")
	source := flag.String("source", "localhost", "The address of the OPC server")
	progID := flag.String("progid", "IntegrationObjects.AdvancedSimulator.1", "The OPC server prog id")
	config := flag.String("c", "", "Configuration file, for example declaring tags to process. If not specified, all tags will be processed")
	create := flag.Bool("create", false, "Use this parameter to create a config file with all tags found in the specified server")
	list := flag.Bool("list", false, "Lists the OPC DA servers available on the specified source")
	browse := flag.String("browse", "", "Lists all tags at the specified tag")
	iterations := flag.Int("i", 1, "Number of times to get all specified tags (used to measure performance)")
	interval := flag.Int("p", 1, "Read interval in seconds")
	flag.Parse()

	if *list {
		if ao, err := opc.NewAutomationObject(); ao != nil && err == nil {
			servers_found := ao.GetOPCServers(*source)
			log.Printf("Found %d server(s) on '%s':\n", len(servers_found), *source)
			for _, server := range servers_found {
				log.Println(server)
			}
		} else {
			log.Println("Unable to get new automation object, err:", err)
		}
		return
	}

	tags := []string{""}
	if *config == "" || *create {
		browser, err := opc.CreateBrowser(*progID, // ProgId
			[]string{*source}, // Nodes
		)
		if err != nil {
			fmt.Println("Failed to browse OPC server, error:", err)
			return
		}

		fmt.Println("Available tags")
		if *browse != "" {
			subtree := opc.ExtractBranchByName(browser, *browse)
			if subtree == nil {
				log.Println("No tags available at specified branch:", *browse)
				return
			}

			opc.PrettyPrint(subtree)
			tags = opc.CollectTags(subtree)
			if *config == "" || !*create {
				return
			}
		} else {
			opc.PrettyPrint(browser)
			tags = opc.CollectTags(browser)
		}
	}

	// Check config and use tags defined there if any
	if *config != "" {
		if !*create {
			if data, err := ioutil.ReadFile(*config); err == nil {
				c := &Config{}
				if err = json.Unmarshal(data, &c); err == nil && len(c.Tags) > 0 {
					tags = make([]string, len(c.Tags))
					for i, tag := range c.Tags {
						tags[i] = tag.Name
					}
				} else {
					log.Printf("Unable to interpret config file, %s. Will be using all tags found in the server. Error: %v\n", *config, err)
				}
			} else {
				log.Printf("Unable to open specified config file, %s. Will be using all tags found in the server. Error: %v\n", *config, err)
			}
		} else {
			c := &Config{}
			c.Tags = make([]ConfigTagEntry, len(tags))
			for t, tag := range tags {
				c.Tags[t].Name = tag
			}
			if data, err := json.MarshalIndent(c, "", "    "); err == nil {
				if err = ioutil.WriteFile(*config, data, 0644); err != nil {
					log.Printf("Unable to write JSON config file, %s. Error: %v\n", *config, err)
				} else {
					log.Println("New configuration file with all tags created:", *config)
					return
				}
			} else {
				log.Printf("Unable to create JSON for config file, %s. Error: %v\n", *config, err)
			}
		}
	}

	if len(tags) <= 0 {
		log.Println("No tags defined, aborting...")
		return
	}

	client, err := opc.NewConnection(*progID, // ProgId
		[]string{*source}, //  OPC servers nodes
		tags,              // slice of OPC tags
	)
	defer client.Close()

	if err != nil {
		fmt.Println("Failed to connect to OPC server, error:", err)
		return
	}

	con, err := net.Dial("udp", fmt.Sprintf("%s:%d", *target, *port))
	if con == nil {
		fmt.Println("Failed to connect target server:", err)
		return
	}
	
	defer con.Close()

	fmt.Println("Connected...")
	timer := time.NewTicker(time.Duration(*interval) * time.Second)

	// read all added tags
	go func() {
		items := client.Read() // This is only to get the number of items
		msg := &DataMessage{}
		msg.Count = 10
		msg.Points = make([]DataPoint, msg.Count)

		for {
			<-timer.C
			start := float64(time.Now().UnixNano())
			for j := 0; j < *iterations; j++ {
				items = client.Read()

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
			if *iterations > 1 {
				fmt.Printf("It took %f seconds to read %d tags\n", (end-start)/1000000000.0, len(items)**iterations)
			}
		}
	}()

	// Sleep until interrupted
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Exiting ...")
	client.Close()
}
