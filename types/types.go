package types

import (
	"time"
)

type NodeUsageType struct {
	ID         uint               `gorm:"primaryKey" json:"id"`
	Percentage int                `json:"percentage"`
	NodeInfo   map[string]float64 `gorm:"type:jsonb" json:"node_info"`
	CreatedAt  time.Time          `json:"created"`
	UpdatedAt  time.Time          `json:"updated"`
}

type NodeInfo struct {
	NodeName  string  `json:"node_name"`
	NodeUsage float64 `json:"usage"`
}

type NodeDrainResult struct {
	NodeName        string  `json:"node_name"`
	InstanceType    string  `json:"instance_type"`
	ProvisionerName string  `json:"provisioner_name"`
	Percentage      float64 `json:"percentage"`
}
