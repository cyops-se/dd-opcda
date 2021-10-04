package engine

import (
	"dd-opcda/db"
	"dd-opcda/types"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/cyops-se/opc"
)

func metaSender(diodeProxy *types.DiodeProxy) {
	address := fmt.Sprintf("%s:%d", diodeProxy.EndpointIP, diodeProxy.MetaPort)

	db.Log("trace", "Setting up outgoing META", address)

	con, err := net.Dial("udp", address)
	if con == nil {
		db.Log("error", "Failed to open emitter", fmt.Sprintf("UDP emitter to IP: %s could not be opened, error: %s", address, err.Error()))
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
				log.Println("Failed to send meta data:", err.Error())
			} else {
				// log.Println("Sending meta data ... ", len(tags), n)
			}
		}

		<-timer.C
	}
}

func groupDataCollector(group *types.OPCGroup, tags []*types.OPCTag) {
	timer := time.NewTicker(time.Duration(group.Interval) * time.Second)

	tagnames := make([]string, len(tags))
	for i, tag := range tags {
		tagnames[i] = tag.Name
	}

	client, err := opc.NewConnection(group.ProgID, // ProgId
		[]string{"localhost"}, //  OPC servers nodes
		tagnames,              // slice of OPC tags
	)
	defer client.Close()

	if err != nil {
		db.Log("error", "Failed to create new connection", fmt.Sprintf("Group name: %s (id: %d), progid: %s, err: %s, %s",
			group.Name, group.ID, group.ProgID, err.Error(), tagnames))
		return
	}

	target := fmt.Sprintf("%s:%d", group.DiodeProxy.EndpointIP, group.DiodeProxy.DataPort)

	db.Log("trace", "Setting up outgoing DATA", target)

	con, err := net.Dial("udp", target)
	if con == nil {
		db.Log("error", "Failed to open emitter", fmt.Sprintf("UDP emitter to IP: %s could not be opened, error: %s", target, err.Error()))
		return
	}

	db.Log("trace", "Collecting tags", fmt.Sprintf("%d tags from group: %s (id: %d)", len(tags), group.Name, group.ID))

	items := client.Read() // This is only to get the number of items
	msg := &types.DataMessage{Version: 2, Group: group.Name, Interval: group.Interval}
	msg.Count = 10
	msg.Points = make([]types.DataPoint, msg.Count)

	// Initiate group running state
	group.Counter = 0
	group.Status = types.GroupStatusRunning
	db.DB.Save(group)

	var i, b int // golang always initialize to 0
	for {
		if g, _ := GetGroup(group.ID); g != nil && g.Status == types.GroupStatusNotRunning {
			db.Log("info", "OPC group stopped", fmt.Sprintf("Group interval timer STOPPED, group: %s (id: %d)", group.Name, group.ID))
			break
		}

		items = client.Read()

		for k, v := range items {
			msg.Points[b].Time = v.Timestamp
			msg.Points[b].Name = k
			msg.Points[b].Value = v.Value
			msg.Points[b].Quality = int(v.Quality)

			// Send batch when msg.Points is full (keep it small to avoid fragmentation)
			if b == len(msg.Points)-1 {
				data, _ := json.Marshal(msg)
				con.Write(data)
				b = 0
				msg.Sequence++
				// log.Printf("Sent data over UDP to '%s', sequence: %d\n", target, msg.Sequence-1)
				NotifySubscribers("info.data", fmt.Sprintf("Sent data over UDP to '%s', sequence: %d\n", target, msg.Sequence-1))
			} else {
				b++
			}
			i++
		}

		group.LastRun = time.Now()
		group.Counter = group.Counter + uint(len(items))

		db.DB.Model(&group).Updates(types.OPCGroup{LastRun: group.LastRun, Counter: group.Counter})
		NotifySubscribers("data.group", group)

		<-timer.C
	}
}

func InitGroups() {
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
		go metaSender(proxy)
	}
}

func GetGroups() ([]*types.OPCGroup, error) {
	var items []*types.OPCGroup
	db.DB.Table("opc_groups").Order("id").Preload("DiodeProxy").Find(&items)
	return items, nil
}

func GetGroup(id uint) (*types.OPCGroup, error) {
	var item types.OPCGroup
	if err := db.DB.Table("opc_groups").Preload("DiodeProxy").Take(&item, id).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

func GetGroupTags(id uint) ([]*types.OPCTag, error) {
	var items []*types.OPCTag
	if err := db.DB.Table("opc_tags").Find(&items, "groupid = ?", id).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func Start(group *types.OPCGroup) (err error) {
	// Make sure the group is not already running
	if group.Status != types.GroupStatusNotRunning {
		err = fmt.Errorf("Group already running, group: %s (id: %d)", group.Name, group.ID)
		db.Log("error", "OPC collection start failed", err.Error())
		return
	}

	var tags []*types.OPCTag
	db.DB.Table("opc_tags").Find(&tags, "group_id = ?", group.ID)
	if len(tags) <= 0 {
		err = fmt.Errorf("Group does not have any tags defined, group: %s (id: %d)", group.Name, group.ID)
		db.Log("error", "OPC collection start failed", err.Error())
		return
	}

	go groupDataCollector(group, tags)

	return
}

func Stop(group *types.OPCGroup) (err error) {
	// Make sure the group is running
	if group.Status != types.GroupStatusRunning {
		err = fmt.Errorf("Group not running, group: %s (id: %d)", group.Name, group.ID)
		db.Log("error", "OPC collection stop failed", err.Error())
		return
	}

	// Stop collection go routine (unsafe)
	group.Status = types.GroupStatusNotRunning
	db.DB.Save(group)

	return
}

func GetTagNames() ([]string, error) {
	var items []string
	if err := db.DB.Table("opc_tags").Where("deleted_at is null").Pluck("Name", &items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func GetTagInfos() (items []*types.TagsInfos, err error) {
	if err = db.DB.Table("opc_tags").Where("deleted_at is null").Find(&items).Error; err != nil {
		return nil, err
	}

	var name = items[len(items)-1].Name
	for i := 1000; i < 10000; i++ {
		newitem := &types.TagsInfos{ID: uint(i), Name: fmt.Sprintf("%s/%d", name, i)}
		items = append(items, newitem)
	}

	return items, nil
}
