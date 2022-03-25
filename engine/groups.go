package engine

import (
	"dd-opcda/db"
	"dd-opcda/logger"
	"dd-opcda/types"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/cyops-se/opc"
)

var opcmutex sync.Mutex // Issue #3, no time to find out where thread insafety is (looks like it's in or below oleutil)

func metaSender(diodeProxy *types.DiodeProxy) {
	defer handlePanic("metaSender")
	address := fmt.Sprintf("%s:%d", diodeProxy.EndpointIP, diodeProxy.MetaPort)

	logger.Log("trace", "Setting up outgoing META", address)

	con, err := net.Dial("udp", address)
	if con == nil {
		logger.Error("Groups engine", "UDP emitter to IP: %s could not be opened, error: %s", address, err.Error())
		return
	}

	timer := time.NewTicker(10 * time.Minute)
	for {
		tags, _ := GetTagInfos()
		batchsize := 100
		for i := 0; i < len(tags); i += batchsize {
			if i+batchsize > len(tags) {
				batchsize = len(tags) - i
			}
			msg, _ := json.Marshal(tags[i : i+batchsize])
			if _, err := con.Write(msg); err != nil {
				logger.Error("Groups engine", "Failed to send meta data: %s", err.Error())
			} else {
				// log.Println("Sending meta data ... ", len(tags), n)
			}
		}

		<-timer.C
	}
}

func read(client *opc.Connection) map[string]opc.Item {
	defer handlePanic("read")
	opcmutex.Lock()
	defer opcmutex.Unlock()
	return (*client).Read()
}

func groupDataCollector(group *types.OPCGroup, tags []*types.OPCTag) {
	defer handlePanic("groupDataCollector")
	timer := time.NewTicker(time.Duration(group.Interval) * time.Second)

	client, err := opc.NewConnectionWithoutTags(group.ProgID, // ProgId
		[]string{"localhost"}, //  OPC servers nodes
	)

	if err != nil {
		logger.Log("error", "Failed to connect OPC DA server", fmt.Sprintf("Group: %s, progid: %s, err: %s", group.Name, group.ProgID, err.Error()))
		logger.NotifySubscribers("group.failed", group)
		return
	}

	defer client.Close()

	// Adding items
	for _, tag := range tags {
		if err := client.AddSingle(tag.Name); err != nil {
			logger.Log("warning", "Unable to collect tag", fmt.Sprintf("%s, group: %s, progid: %s", tag.Name, group.Name, group.ProgID))
		}
	}

	if len(client.Tags()) == 0 {
		logger.Log("error", "No tags to collect", fmt.Sprintf("Group: %s, progid: %s", group.Name, group.ProgID))
		logger.NotifySubscribers("group.failed", group)
		return
	}

	logger.Log("trace", "Collecting tags", fmt.Sprintf("%d tags from group: %s", len(client.Tags()), group.Name))

	items := read(&client) // This is only to get the number of items
	msg := &types.DataMessage{Version: 2, Group: group.Name, Interval: group.Interval}
	msg.Count = 10
	msg.Points = make([]types.DataPoint, msg.Count)

	// Initiate group running state

	if len(client.Tags()) != len(tags) {
		group.Status = types.GroupStatusRunningWithWarning
		logger.NotifySubscribers("group.warning", group)
	} else {
		group.Status = types.GroupStatusRunning
		logger.NotifySubscribers("group.started", group)
	}

	group.Counter = 0
	db.DB.Save(group)

	var i, b int // golang always initialize to 0
	for {
		if g, _ := GetGroup(group.ID); g != nil && g.Status == types.GroupStatusNotRunning {
			logger.Log("info", "OPC group stopped", fmt.Sprintf("Group interval timer STOPPED, group: %s", group.Name))
			logger.NotifySubscribers("group.stopped", group)
			break
		}

		items = read(&client)

		for k, v := range items {
			msg.Points[b].Time = v.Timestamp
			msg.Points[b].Name = k
			msg.Points[b].Value = v.Value
			msg.Points[b].Quality = int(v.Quality)

			// Send batch when msg.Points is full (keep it small to avoid fragmentation)
			if b == len(msg.Points)-1 {
				data, _ := json.Marshal(msg)
				if proxy := FirstProxy(); proxy != nil {
					proxy.DataChan <- data
					b = 0
					msg.Sequence++
					logger.NotifySubscribers("data.message", string(data))
					cacheMessage(msg)
				}
			} else {
				b++
			}
			i++
		}

		group.LastRun = time.Now()
		group.Counter = group.Counter + uint(len(items))

		db.DB.Model(&group).Updates(types.OPCGroup{LastRun: group.LastRun, Counter: group.Counter})
		logger.NotifySubscribers("data.group", group)

		<-timer.C
	}
}

func InitGroups() {
	defer handlePanic("InitGroups")
	InitSetting("tagpathdelimiter", ".", "Delimiter in OPC DA tag paths. Differs between OPC DA servers")

	items, _ := GetGroups()
	for _, item := range items {
		item.Status = types.GroupStatusNotRunning
		db.DB.Save(item)

		if item.RunAtStart {
			Start(item)
		}
	}

	var proxies []*types.DiodeProxy
	db.DB.Table("diode_proxies").Order("id").Find(&proxies)
	for _, proxy := range proxies {
		initProxy(proxy)
		// go metaSender(proxy)
	}
}

func GetGroups() ([]*types.OPCGroup, error) {
	defer handlePanic("GetGroups")
	var items []*types.OPCGroup
	db.DB.Table("opc_groups").Order("id").Preload("DiodeProxy").Find(&items)
	return items, nil
}

func GetGroup(id uint) (*types.OPCGroup, error) {
	defer handlePanic("GetGroup")
	var item types.OPCGroup
	if err := db.DB.Table("opc_groups").Preload("DiodeProxy").Take(&item, id).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

func GetDefaultGroup() (*types.OPCGroup, error) {
	defer handlePanic("GetDefaultGroup")
	var item types.OPCGroup
	if err := db.DB.Table("opc_groups").Preload("DiodeProxy").First(&item, "default_group = true").Error; err != nil {
		return nil, err
	}

	return &item, nil
}

func GetGroupTags(id uint) ([]*types.OPCTag, error) {
	defer handlePanic("GetGroupTags")
	var items []*types.OPCTag
	if err := db.DB.Table("opc_tags").Find(&items, "groupid = ?", id).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func Start(group *types.OPCGroup) (err error) {
	defer handlePanic("Start")
	// Make sure the group is not already running
	if group.Status != types.GroupStatusNotRunning {
		err = fmt.Errorf("Group already running, group: %s (id: %d)", group.Name, group.ID)
		logger.Log("error", "OPC collection start failed", err.Error())
		return
	}

	var tags []*types.OPCTag
	db.DB.Table("opc_tags").Find(&tags, "group_id = ?", group.ID)
	if len(tags) <= 0 {
		err = fmt.Errorf("Group does not have any tags defined, group: %s (id: %d)", group.Name, group.ID)
		logger.Log("error", "OPC collection start failed", err.Error())
		return
	}

	go groupDataCollector(group, tags)

	return
}

func Stop(group *types.OPCGroup) (err error) {
	defer handlePanic("Stop")
	// Make sure the group is running
	if group.Status != types.GroupStatusRunning && group.Status != types.GroupStatusRunningWithWarning {
		err = fmt.Errorf("Group not running, group: %s (id: %d)", group.Name, group.ID)
		logger.Log("error", "OPC collection stop failed", err.Error())
		return
	}

	// Stop collection go routine (unsafe)
	group.Status = types.GroupStatusNotRunning
	db.DB.Save(group)

	return
}

func GetTagNames() ([]string, error) {
	defer handlePanic("GetTagNames")
	var items []string
	if err := db.DB.Table("opc_tags").Where("deleted_at is null").Pluck("Name", &items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func GetTagInfos() (items []*types.TagsInfos, err error) {
	defer handlePanic("GetTagInfos")
	if err = db.DB.Table("opc_tags").Where("deleted_at is null").Find(&items).Error; err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("Could not find any tags")
	}

	return items, nil
}
