package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
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
	Count  int         `json:"count"`
	Points []DataPoint `json:"points"`
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
	flag.Parse()

	if *list {
		servers_found := opc.NewAutomationObject().GetOPCServers(*source)
		log.Printf("Found %d server(s) on '%s':\n", len(servers_found), *source)
		for _, server := range servers_found {
			log.Println(server)
		}
		return
	}

	con, err := net.Dial("udp", fmt.Sprintf("%s:%d", *target, *port))
	defer con.Close()

	browser, err := opc.CreateBrowser(*progID, // ProgId
		[]string{*source}, // Nodes
	)
	if err != nil {
		fmt.Println("Failed to browse OPC server, error:", err)
		return
	}

	fmt.Println("Available tags")
	opc.PrettyPrint(browser)
	tags := opc.CollectTags(browser)

	// Check config and use tags defined there if any
	if *config != "" {
		if !*create {
			if data, err := ioutil.ReadFile(*config); err == nil {
				c := &Config{}
				if err = json.Unmarshal(data, &c); err == nil && len(c.Tags) > 0 {
					tags := make([]string, len(c.Tags))
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
			for i, tag := range tags {
				c.Tags[i].Name = tag
			}
			if data, err := json.MarshalIndent(c, "", "    "); err == nil {
				if err = ioutil.WriteFile(*config, data, 0644); err != nil {
					log.Printf("Unable to write JSON config file, %s. Error: %v\n", *config, err)
				} else {
					return
				}
			} else {
				log.Printf("Unable to create JSON for config file, %s. Error: %v\n", *config, err)
			}
		}
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

	fmt.Println("Connected...")
	timer := time.NewTicker(1 * time.Second)

	// read all added tags
	go func() {
		items := client.Read()
		msg := &DataMessage{}
		msg.Count = len(items)
		msg.Points = make([]DataPoint, msg.Count)

		for {
			<-timer.C
			items = client.Read()
			i := 0
			for k, v := range items {
				msg.Points[i].Time = v.Timestamp
				msg.Points[i].Name = k
				msg.Points[i].Value = v.Value
				msg.Points[i].Quality = int(v.Quality)
				i++
			}

			data, _ := json.Marshal(msg)
			con.Write(data)
			// fmt.Printf("%d bytes written\n", n)
			// fmt.Println(string(data))
		}
	}()

	// Sleep forever
	<-(chan int)(nil)
}
