package engine

import (
	"dd-opcda/db"
	"dd-opcda/types"
	"fmt"
)

func GetSetting(key string) (types.KeyValuePair, error) {
	var item types.KeyValuePair
	err := db.DB.Table("key_value_pairs").Find(&item, "key = ?", key).Error
	return item, err
}

func PutSetting(key string, value string) {
	if item, err := GetSetting(key); err == nil {
		item.Value = value
		db.DB.Save(item)
	} else {
		item := types.KeyValuePair{Key: key, Value: value}
		db.DB.Create(item)
	}
}

func InitSetting(key string, value string) {
	if _, err := GetSetting(key); err != nil {
		fmt.Println("SETTING:", key, "initalized with", value)
		PutSetting(key, value)
	}
}
