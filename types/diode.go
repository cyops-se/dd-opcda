package types

import "gorm.io/gorm"

type DiodeProxy struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	EndpointIP  string `json:"ip"`
	EndpointMAC string `json:"mac"`
	MetaPort    int    `json:"metaport"`
	DataPort    int    `json:"dataport"`
	FilePort    int    `json:"fileport"`
}
