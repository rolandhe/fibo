package fibo

import (
	"time"
)

type zookeeperConf struct {
	servers        []string
	sessionTimeout time.Duration
	maxWorkerBits  int
	maxIdcBits     int
}

type configure struct {
	maxWorkerBits int
	maxIdcBits    int
	nameSpaces    []string
}

func getZkConf() *zookeeperConf {
	return &zookeeperConf{
		servers:        []string{"localhost:2181", "localhost:2182", "localhost:2183"},
		sessionTimeout: time.Second * 2,
		maxWorkerBits:  3,
		maxIdcBits:     3,
	}
}

func getConfigure() *configure {

	return &configure{
		maxWorkerBits: 3,
		maxIdcBits:    3,
		nameSpaces:    []string{"default"},
	}
}
