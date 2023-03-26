package core

import (
	"errors"
	"github.com/go-zookeeper/zk"
	"github.com/rolandhe/fibo/logger"
	"strconv"
	"time"
)

var workerPath = "/core/workers"
var workerName = workerPath + "/worker"

func reconnect(conf *zookeeperConf, g *Generator) {
	for i := 0; i < 5; i++ {
		err := initWorkerCore(conf, g)
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 1000)
		logger.GLogger.Infoln("init workerId failed,try again")
	}
}

func InitWorkerId(g *Generator) error {
	conf := getZkConf()
	return initWorkerCore(conf, g)
}

func initWorkerCore(conf *zookeeperConf, g *Generator) error {
	var conn *zk.Conn

	conn, eventCh, err := zk.Connect(conf.servers, conf.sessionTimeout)
	if err != nil {
		logger.GLogger.Infoln(err)
		return err
	}

	workerId, err := getWorkerId(conn, conf)
	if err != nil {
		return err
	}
	g.SetWorkerId(workerId)
	g.SetState(true)
	logger.GLogger.Infof("init worker id is %d\n", workerId)
	go func() {
		for {
			select {
			case event := <-eventCh:
				if event.State == zk.StateDisconnected {
					logger.GLogger.Infoln("waring: connect is disconnect, next to reconnect")
					conn.Close()
					reconnect(conf, g)
					return
				}
				logger.GLogger.Infoln(event)
			case <-time.After(time.Minute):
				logger.GLogger.Infoln("can't get event, wait minute")
			}
		}
	}()
	return nil
}

func getWorkerId(conn *zk.Conn, conf *zookeeperConf) (int32, error) {
	var err error
	if err = ensureNode(conn, "/core"); err != nil {
		conn.Close()
		return 0, err
	}

	if err = ensureNode(conn, "/core/workers"); err != nil {
		conn.Close()
		return 0, err
	}

	locker := zk.NewLock(conn, "/core/lock", zk.WorldACL(zk.PermAll))
	err = locker.Lock()
	if err != nil {
		logger.GLogger.Infoln(err)
		conn.Close()
		return 0, err
	}
	defer locker.Unlock()
	maxWorkers := int32(1 << conf.maxWorkerBits)

	existsWorkers, _, err := conn.Children(workerPath)

	if err != nil {
		conn.Close()
		return 0, err
	}

	if len(existsWorkers) >= int(maxWorkers) {
		logger.GLogger.Infoln("workers enough")
		conn.Close()
		return 0, errors.New("workers enough")
	}
	var existsId []int32
	for _, existName := range existsWorkers {
		existsId = append(existsId, parseIdByName(existName, maxWorkers))
	}

	for {
		name, err := conn.Create(workerName, nil, zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
		if err != nil {
			conn.Close()
			return 0, err
		}
		workerId := parseIdByName(name, maxWorkers)
		if !repeat(existsId, workerId) {
			return workerId, nil
		}
		err = conn.Delete(name, 0)
		if err != nil {
			conn.Close()
			return 0, err
		}
	}
}

func repeat(existIds []int32, workerId int32) bool {
	for _, v := range existIds {
		if v == workerId {
			return true
		}
	}
	return false
}

func parseIdByName(name string, maxWorkers int32) int32 {
	l := len(workerPath)
	if l > len(name) {
		logger.GLogger.Infoln("error", name)
		panic(name)
	}
	number := name[l:]
	id, _ := strconv.Atoi(number)
	return int32(int64(id) % int64(maxWorkers))
}

func ensureNode(conn *zk.Conn, path string) error {
	_, stat, err := conn.Get(path)
	if err == nil {
		logger.GLogger.Infoln(path, stat.Czxid)
		return nil
	}
	if err != nil && err != zk.ErrNoNode {
		return err
	}

	name, err := conn.Create(path, nil, 0, zk.WorldACL(zk.PermAll))
	if err == zk.ErrNodeExists {
		logger.GLogger.Infoln(err)
		return nil
	}
	if err == nil {
		logger.GLogger.Infoln(name)
	}

	return err
}
