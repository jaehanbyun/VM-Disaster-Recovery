package model

import "github.com/jaehanbyun/VM-Disaster-Recovery/data"

type DBHandler interface {
	GetWeight() (data.Weight, error)
	SetWeight(data.Weight) error
	GetThreshold() (float32, error)
	SetThreshold(float32) error
	Close()
}

func NewDBHandler() DBHandler {
	return newPostgresHandler()
}
