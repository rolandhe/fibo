package fibo

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

const maxIncrId = 4095
const MaxBatchSize = 8192
const IncrIdBits = 12

var err500 = errors.New("internal error")
var errNoNs = errors.New("invalid id namespace")
var errExceedBatch = errors.New("exceed max batch size")
var errBatchSize = errors.New("invalid batch size")

func NewGenerator() *Generator {
	conf := getConfigure()
	nsMap := make(map[string]*NameSpace)
	for _, ns := range conf.nameSpaces {
		nsMap[ns] = &NameSpace{Name: ns}
	}
	return &Generator{
		workerIdBits: conf.maxWorkerBits,
		idcBits:      conf.maxIdcBits,
		NameSpaces:   nsMap,
	}
}

type BatchIds struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type Generator struct {
	state        atomic.Bool
	workerId     atomic.Int32
	idcId        int
	workerIdBits int
	idcBits      int

	NameSpaces map[string]*NameSpace
}

type NameSpace struct {
	sync.Mutex
	Name      string
	timeStamp int64
	nextId    int64
}

func (g *Generator) SetState(state bool) {
	g.state.Store(state)
}
func (g *Generator) SetWorkerId(workerId int32) {
	g.workerId.Store(workerId)
}

func (g *Generator) GenOneId(name string) (int64, error) {
	if !g.state.Load() {
		return 0, err500
	}
	nameSpace, ok := g.NameSpaces[name]
	if !ok {
		return 0, errNoNs
	}

	nameSpace.Lock()
	defer nameSpace.Unlock()
	curTime := time.Now().UnixMilli()
	if curTime != nameSpace.timeStamp {
		nameSpace.timeStamp = curTime
		nameSpace.nextId = 0
	}
	if nameSpace.nextId < maxIncrId {
		id := g.composeId(nameSpace, nameSpace.nextId)
		nameSpace.nextId++
		return id, nil
	}
	nano := time.Now().UnixNano()
	sleep := (nameSpace.timeStamp+1)*1000000 - nano
	nameSpace.nextId = 0
	if sleep > 0 {
		time.Sleep(time.Duration(sleep))
		nameSpace.timeStamp++
	} else {
		nameSpace.timeStamp = nano / 1000000
	}

	id := g.composeId(nameSpace, nameSpace.nextId)
	nameSpace.nextId++
	return id, nil
}

func (g *Generator) GenBatchId(name string, batch int64) ([]*BatchIds, error) {
	if batch <= 0 {
		return nil, errBatchSize
	}
	if batch > MaxBatchSize {
		return nil, errExceedBatch
	}
	var ret []*BatchIds
	if batch == 1 {
		id, err := g.GenOneId(name)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &BatchIds{id, id})
		return ret, nil
	}
	if !g.state.Load() {
		return nil, err500
	}
	nameSpace, ok := g.NameSpaces[name]
	if !ok {
		return nil, errNoNs
	}
	nameSpace.Lock()
	defer nameSpace.Unlock()
	curTime := time.Now().UnixMilli()
	if curTime != nameSpace.timeStamp {
		nameSpace.timeStamp = curTime
		nameSpace.nextId = 0
	}

	if maxIncrId+1-nameSpace.nextId >= batch {
		start := nameSpace.nextId
		nameSpace.nextId += batch
		ret = append(ret, &BatchIds{start, nameSpace.nextId - 1})
		return ret, nil
	}
	ret = append(ret, &BatchIds{g.composeId(nameSpace, nameSpace.nextId), g.composeId(nameSpace, maxIncrId)})
	batch -= maxIncrId - nameSpace.nextId + 1
	waitMs := (batch + maxIncrId) / (maxIncrId + 1)
	sleep := (nameSpace.timeStamp+waitMs)*1000000 - time.Now().UnixNano()
	nameSpace.nextId = 0
	if sleep > 0 {
		time.Sleep(time.Duration(sleep))
	}

	for batch > 0 {
		nameSpace.timeStamp++
		if batch >= maxIncrId+1 {
			ret = append(ret, &BatchIds{g.composeId(nameSpace, 0), g.composeId(nameSpace, maxIncrId)})
			batch -= maxIncrId + 1
			continue
		}
		ret = append(ret, &BatchIds{g.composeId(nameSpace, 0), g.composeId(nameSpace, batch-1)})
		nameSpace.nextId = batch
		batch = 0
	}
	return ret, nil
}

func (g *Generator) composeId(nameSpace *NameSpace, v int64) int64 {
	id := nameSpace.timeStamp << (g.idcBits + g.workerIdBits + IncrIdBits)
	id |= int64(g.idcId) << (g.workerIdBits + IncrIdBits)
	id |= int64(g.workerId.Load()) << IncrIdBits
	id |= v
	return id
}
