package model

import "github.com/jaehanbyun/VM-Disaster-Recovery/data"

type DBHandler interface {
	Close()
	Init() error
	GetWeight() (data.Weight, error)
	SetWeight(data.Weight) error
	GetThreshold() (float32, error)
	SetThreshold(float32) error
	GetVMInfo(string) (*data.VMInstance, error)
	GetVMsInfo() ([]*data.VMInstance, error)
	SetVMInfo(data.VMInstance) error
	SetVMsInfo() error
	GetImageName(string) (string, error)
}

func NewDBHandler() DBHandler {
	return newPostgresHandler()
}
