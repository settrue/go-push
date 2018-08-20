package gateway

import (
	"github.com/settrue/go-push/common"
	"sync"
)

var (
	PUSH_ALL int = 0
	PUSH_GROUP int = 1
)

type Job struct {
	Type int
	GroupId uint64
	Msgs *common.WSMsg
}

type Bucket struct{
	Id uint64
	groupMap map[uint64]*Group
	connMap map[uint64]*WSConn
	jobChan chan*Job
	rwMutex sync.RWMutex
}

func InitBucket(id uint64) *Bucket{
	bucket := &Bucket{
		Id:id,
		groupMap:make(map[uint64]*Group),
		connMap:make(map[uint64]*WSConn),
		jobChan:make(chan*Job),
	}

	go bucket.workLoop()
	return bucket
}

func (self *Bucket)PushAll(msg *common.WSMsg) {
	var (
		conn *WSConn
	)
	self.rwMutex.RLock()
	defer self.rwMutex.RUnlock()

	for _, conn = range self.connMap {
		conn.Write(msg)
	}
}

func (self *Bucket)PushGroup(id uint64, msg *common.WSMsg) {
	var (
		group *Group
		exists bool
	)
	self.rwMutex.RLock()
	group, exists = self.groupMap[id]
	self.rwMutex.RUnlock()
	if !exists {
		return
	}

	group.Push(msg)
}

func (self *Bucket)workLoop(){
	var (
		job *Job
	)
	for {
		select{
		case job = <- self.jobChan:
			if job.Type == PUSH_ALL {
				self.PushAll(job.Msgs)
			} else {
				self.PushGroup(job.GroupId, job.Msgs)
			}
		}
	}
}

func (self *Bucket)Write(job *Job) {
	select {
	case self.jobChan <- job:
	}
	return
}