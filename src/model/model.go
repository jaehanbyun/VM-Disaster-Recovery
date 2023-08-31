package model

import "github.com/jaehanbyun/VM-Disaster-Recovery/data"

type DBHandler interface {
	GetWeight() (data.Weight, error)
	SetWeight(data.Weight) error
	Close()
}

func NewDBHandler() DBHandler {
	return newPostgresHandler()
}
