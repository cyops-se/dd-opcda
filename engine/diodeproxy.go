package engine

import (
	"dd-opcda/db"
	"dd-opcda/types"
	"fmt"
	"log"
	"net"
)

var proxies map[uint]*types.DiodeProxy

func initProxy(proxy *types.DiodeProxy) (err error) {
	// Initialize channels and UDP sinks

	if proxies == nil {
		proxies = map[uint]*types.DiodeProxy{}
	}

	// DATA
	target := fmt.Sprintf("%s:%d", proxy.EndpointIP, proxy.DataPort)
	proxy.DataCon, err = net.Dial("udp", target)
	if proxy.DataCon == nil {
		db.Log("error", "Failed to open data emitter", fmt.Sprintf("UDP data emitter to IP: %s could not be opened, error: %s", target, err.Error()))
	} else {
		db.Log("trace", "Setting up outgoing DATA", target)
		proxy.DataChan = make(chan []byte)
		go sendJob(proxy.DataChan, proxy.DataCon)
	}

	// META
	target = fmt.Sprintf("%s:%d", proxy.EndpointIP, proxy.MetaPort)
	proxy.MetaCon, err = net.Dial("udp", target)
	if proxy.MetaCon == nil {
		db.Log("error", "Failed to open meta emitter", fmt.Sprintf("UDP meta emitter to IP: %s could not be opened, error: %s", target, err.Error()))
	} else {
		db.Log("trace", "Setting up outgoing META", target)
		proxy.MetaChan = make(chan []byte)
		go sendJob(proxy.MetaChan, proxy.MetaCon)
	}

	// FILES
	target = fmt.Sprintf("%s:%d", proxy.EndpointIP, proxy.FilePort)
	proxy.FileCon, err = net.Dial("udp", target)
	if proxy.FileCon == nil {
		db.Log("error", "Failed to open file emitter", fmt.Sprintf("UDP file emitter to IP: %s could not be opened, error: %s", target, err.Error()))
	} else {
		db.Log("trace", "Setting up outgoing FILE", target)
		proxy.MetaChan = make(chan []byte)
		go sendJob(proxy.FileChan, proxy.FileCon)
	}

	fmt.Printf("PROXY: proxy with ID %d initialized\n", proxy.ID)
	proxies[proxy.ID] = proxy

	return err
}

func sendJob(channel chan []byte, connection net.Conn) {
	for {
		data := <-channel
		if _, err := connection.Write(data); err != nil {
			log.Printf("PROXY: Failed to send %#v, error: %s\n", data, err.Error())
		}
	}
}
