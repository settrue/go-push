package gateway

import (
	"time"
	"sync"
)

type BatchMsg struct{
	items []*Msg
	timer *time.Timer
}

type Msg struct{
	data []byte
}

func InitMergerWorker(groupId uint64) (mw *MergerWorker) {
	mw = &MergerWorker{
		groupId:groupId,
		msgs:make(chan*Msg, conf.mergerMsgNum),
		timeoutMsgs:make(chan*BatchMsg, conf.mergerTimeoutNum),
		closeChan:make(chan byte),
	}
	go mw.mergerWorkerMain()
	return
}

type MergerWorker struct {
	groupId uint64
	msgs chan*Msg
	batchMsg *BatchMsg
	timeoutMsgs chan*BatchMsg
	closeChan chan byte
}

func (self *MergerWorker)mergerWorkerMain() {
	var (
		msg *Msg
		batch *BatchMsg
	)
	for {
		select {
		case msg = <- self.msgs:
			if self.batchMsg == nil {
				self.batchMsg = &BatchMsg{}
				self.batchMsg.timer = time.AfterFunc(time.Duration(conf.msgInterval) * time.Second, func(){
					self.timeoutMsgs <- self.batchMsg
					self.batchMsg = nil
				})
			}
			self.batchMsg.items = append(self.batchMsg.items, msg)
			if len(self.batchMsg.items) > 10 {
				self.batchMsg.timer.Stop()
			}
		case batch = <- self.timeoutMsgs:
			master.Write(batch, self.groupId)
		case <- self.closeChan:
			return
		}
	}
}

func (self *MergerWorker)Close(){
	close(self.closeChan)
	delete(merger.groupWorkers, self.groupId)
}

type Merger struct {
	groupWorkers map[uint64]*MergerWorker
	broadcastWorker *MergerWorker
	mutex sync.Mutex
}

var (
	merger *Merger
)

func InitMerger() {
	merger = &Merger{}
}
func (self *Merger) Write(groupId uint64, data []byte){
	var (
		msg *Msg
		existed bool
		worker *MergerWorker
	)
	msg = &Msg{
		data,
	}
	if worker,existed = self.groupWorkers[groupId]; !existed {
		worker = InitMergerWorker(groupId)
		self.mutex.Lock()
		self.groupWorkers[groupId] = worker
		self.mutex.Unlock()
	}
	worker.Write(msg)
}

func (self *MergerWorker) Write(msg *Msg){
	self.msgs <- msg
}