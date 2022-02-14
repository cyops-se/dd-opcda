package engine

import (
	"dd-opcda/logger"
	"fmt"
	"log"
	"sync"

	"github.com/cyops-se/opc"
	"github.com/go-ole/go-ole"
)

type Server struct {
	ID     int    `json:id`
	ProgID string `json:"progid"`
	Cursor *ole.VARIANT
}

var servers []*Server
var mutex sync.Mutex

func handlePanic() {
	if r := recover(); r != nil {
		log.Println(r)
		return
	}
}

func Lock() {
	mutex.Lock()
}

func Unlock() {
	mutex.Unlock()
}

func InitServers() {
	defer handlePanic()

	i := 0
	if ao := opc.NewAutomationObject(); ao != nil {
		serversfound := ao.GetOPCServers("localhost")
		logger.Log("trace", "OPC server init", fmt.Sprintf("Found %d server(s) on '%s':\n", len(serversfound), "localhost"))
		for _, server := range serversfound {
			logger.Log("trace", "OPC server found", server)
			servers = append(servers, &Server{ProgID: server, ID: i})
			i++
		}
	} else {
		logger.Log("error", "OPC server init failure", "Unable to get new automation object")
	}
}

func GetServers() []*Server {
	return servers
}

func GetServer(sid int) (*Server, error) {
	if sid < 0 || sid >= len(servers) {
		return nil, fmt.Errorf("no such server id: %d", sid)
	}

	return servers[sid], nil
}

func GetBrowser(sid int) (*ole.VARIANT, error) {
	server, err := GetServer(sid)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if server.Cursor == nil {
		mutex.Lock()
		server.Cursor, _ = opc.CreateBrowserCursor(server.ProgID, []string{"localhost"})
		mutex.Unlock()
	}

	return server.Cursor, nil
}
