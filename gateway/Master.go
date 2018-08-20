package gateway

import (
	"github.com/gin-gonic/gin/json"
	"github.com/settrue/go-push/common"
)

type Master struct{
	buckets []*Bucket
	dispatchChan chan*Job
}

var (
	master *Master
)

func InitMaster() {
	master = &Master{
		buckets:make([]*Bucket, conf.bucketsNum),
		dispatchChan:make(chan*Job, conf.dispatchChanNum),
	}

	var (
		i uint64
	)
	for i=0;i<conf.bucketsNum;i++ {
		master.buckets[i] = InitBucket(i)
	}
	for i=0;i<conf.dispatchWorkerNum;i++{
		go master.dispatchWorkerMain()
	}
}

func (self *Master)dispatchWorkerMain() {
	var (
		pushJob *Job
		bucket *Bucket
	)
	for {
		select {
		case pushJob = <- self.dispatchChan:
			for _, bucket = range self.buckets {
				bucket.Write(pushJob)
			}
		}
	}
}

func (self *Master)GetBucket(conn *WSConn) *Bucket {
	return self.buckets[conn.Id % conf.bucketsNum]
}

func (self *Master)Write(b *BatchMsg, groupId uint64) (err error) {
	var (
		bt []byte
		job *Job
		tp int
		wsMsg *common.WSMsg
	)
	if bt, err = json.Marshal(b.items); err != nil {
		return
	}
	if groupId == 0 {
		tp = PUSH_ALL
	} else {
		tp = PUSH_GROUP
	}

	wsMsg = common.MkWSMsg(tp, bt)
	job = &Job{
		tp,
		groupId,
		wsMsg,
	}
	self.dispatchChan <- job
	return
}