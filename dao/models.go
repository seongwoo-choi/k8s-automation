package dao

import (
	"time"
)

type NodeDrainResultDB struct {
	ID         uint   `gorm:"primaryKey"`
	NodeInfo   string `gorm:"type:jsonb"`
	Percentage string
	Quantity   int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type NodeDrainStatusDB struct {
	ID            uint `gorm:"primaryKey"`
	NodeName      string
	Status        string `gorm:"default:pending"` // pending, in_progress, completed, failed
	StartedAt     time.Time
	EndedAt       time.Time
	Error         string
	DrainResultID uint              `gorm:"index"`
	DrainResult   NodeDrainResultDB `gorm:"foreignKey:DrainResultID"`
}

type NodeMemoryUsageDB struct {
	ID         uint   `gorm:"primaryKey"`
	NodeInfo   string `gorm:"type:jsonb"`
	Percentage string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type NodeDiskUsageDB struct {
	ID         uint   `gorm:"primaryKey"`
	NodeInfo   string `gorm:"type:jsonb"`
	Percentage string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

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
