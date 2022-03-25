package engine

import (
	"dd-opcda/logger"
	"dd-opcda/types"
	"fmt"
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
		logger.Log("error", "Failed to open data emitter", fmt.Sprintf("UDP data emitter to IP: %s could not be opened, error: %s", target, err.Error()))
	} else {
		logger.Log("trace", "Setting up outgoing DATA", target)
		proxy.DataChan = make(chan []byte)
		go sendJob(proxy.DataChan, proxy.DataCon)
	}

	// META
	target = fmt.Sprintf("%s:%d", proxy.EndpointIP, proxy.MetaPort)
	proxy.MetaCon, err = net.Dial("udp", target)
	if proxy.MetaCon == nil {
		logger.Log("error", "Failed to open meta emitter", fmt.Sprintf("UDP meta emitter to IP: %s could not be opened, error: %s", target, err.Error()))
	} else {
		logger.Log("trace", "Setting up outgoing META", target)
		proxy.MetaChan = make(chan []byte)
		go sendJob(proxy.MetaChan, proxy.MetaCon)
	}

	// FILES
	target = fmt.Sprintf("%s:%d", proxy.EndpointIP, proxy.FilePort)
	proxy.FileCon, err = net.Dial("udp", target)
	if proxy.FileCon == nil {
		logger.Log("error", "Failed to open file emitter", fmt.Sprintf("UDP file emitter to IP: %s could not be opened, error: %s", target, err.Error()))
	} else {
		logger.Log("trace", "Setting up outgoing FILE", target)
		proxy.MetaChan = make(chan []byte)
		go sendJob(proxy.FileChan, proxy.FileCon)
	}

	logger.Trace("Proxy", "Proxy with ID %d initialized", proxy.ID)
	proxies[proxy.ID] = proxy

	return err
}

func sendJob(channel chan []byte, connection net.Conn) {
	for {
		data := <-channel
		if _, err := connection.Write(data); err != nil {
			logger.Error("Proxy", "Failed to send %#v, error: %s", data, err.Error())
		}
	}
}

func FirstProxy() *types.DiodeProxy {
	for _, p := range proxies {
		return p
	}
	return nil
}
