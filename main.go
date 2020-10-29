package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

func main() {
	target := flag.String("target", "172.16.11.3", "The address of the outer inserter")
	port := flag.Int("port", 4357, "The UDP port of the outer inserter")
	source := flag.String("source", "192.168.0.206", "The address of the OPC server")
	progID := flag.String("progid", "IntegrationObjects.AdvancedSimulator.1", "The OPC server prog id")
	flag.Parse()

	con, err := net.Dial("udp", fmt.Sprintf("%s:%d", *target, *port))
	defer con.Close()

	browser, err := opc.CreateBrowser(*progID, // ProgId
		[]string{*source}, // Nodes
	)
	if err != nil {
		fmt.Println("Failed to browse OPC server, error:", err)
		return
	}

	opc.PrettyPrint(browser)

	client, err := opc.NewConnection(*progID, // ProgId
		[]string{*source},        //  OPC servers nodes
		opc.CollectTags(browser), // slice of OPC tags
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
