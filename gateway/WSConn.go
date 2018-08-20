package gateway

import (
	"github.com/gorilla/websocket"
	"github.com/settrue/go-push/common"
	"time"
	"sync"
)

type WSConn struct{
	Id uint64
	Socket *websocket.Conn
	inChan chan*common.WSMsg
	outChan chan*common.WSMsg
	closeChan chan byte
	lastHeartbeatTime time.Time
	Groups map[uint64]*Group
	mutex sync.Mutex
	isClosed bool
	bucket *Bucket
}

func (self *WSConn)readLoop() {
	var (
		err error
		p []byte
		msg *common.WSMsg
		tp int
	)
	for {
		if tp, p, err = self.Socket.ReadMessage(); err != nil {
			// 读取错误
			self.Close()
			break
		}

		msg = common.MkWSMsg(tp, p)

		select {
		case self.inChan <- msg:
		case <-self.closeChan:
			break
		}
	}
}

func (self *WSConn)writeLoop() {
	var (
		wsMsg *common.WSMsg
		err error
	)
	for {
		select {
		case wsMsg = <-self.outChan:
			if err = self.Socket.WriteMessage(wsMsg.Type, wsMsg.Data); err != nil {
				self.Close()
				break
			}
		case <-self.closeChan:
			break
		}
	}
}

func (self *WSConn)Read() (msg *common.WSMsg, err error){
	select {
	case msg = <- self.inChan:
	case <-self.closeChan:
		err = common.SOCKET_HAVE_CLOSE
	}
	return
}

func (self *WSConn)Write(msg *common.WSMsg) (err error) {
	select {
	case self.outChan <- msg:
	case <-self.closeChan:
		err = common.SOCKET_HAVE_CLOSE
	default:
		err = common.OUT_CHAN_IS_FULL
	}
	return
}

func (self *WSConn)Close() {
	self.Socket.Close()

	self.mutex.Lock()
	defer self.mutex.Unlock()

	if !self.isClosed {
		self.isClosed = true
		close(self.closeChan)
	}
}

func (self *WSConn) IsAlive() bool{
	var (
		now = time.Now()
	)

	self.mutex.Lock()
	defer self.mutex.Unlock()

	if self.isClosed || now.Sub(self.lastHeartbeatTime) > time.Duration(conf.heartbeatInterval) * time.Second {
		return true
	}
	return false
}

func (self *WSConn)KeepAlive() {
	var (
		now = time.Now()
	)

	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.lastHeartbeatTime = now
}

func InitConn(id uint64, socket *websocket.Conn) *WSConn {
	ws := &WSConn{
		Id:id,
		Socket:socket,
		inChan:make(chan*common.WSMsg, conf.inChanNum),
		outChan:make(chan*common.WSMsg, conf.inChanNum),
		closeChan:make(chan byte),
		lastHeartbeatTime:time.Now(),
		Groups:make(map[uint64]*Group),
	}

	ws.joinBucket()

	go ws.readLoop()
	go ws.writeLoop()
	return ws
}

func (self *WSConn)LeaveAll() {
	var (
		exists bool
		id uint64
		group *Group
	)
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if _, exists = self.bucket.connMap[self.Id]; exists {
		self.bucket.rwMutex.Lock()
		delete(self.bucket.connMap, self.Id)
		self.bucket.rwMutex.Unlock()
	}

	for id, group = range self.Groups {
		group.Leave(self)
		//delete(self.Groups, id)
	}
}

func (self *WSConn) LeaveGroup (id uint64) {
	var (
		group *Group
		existed bool
	)
	if group, existed = self.Groups[id]; existed{
		self.mutex.Lock()
		delete(self.Groups, id)
		self.mutex.Unlock()
	} else if group, existed = self.bucket.groupMap[id]; !existed {
		return
	}
	group.Leave(self)

	if group.Count() == 0 {
		self.bucket.rwMutex.Lock()
		delete(self.bucket.groupMap, id)
		self.bucket.rwMutex.Unlock()
	}
}

func (self *WSConn) JoinGroup(id uint64) {
	var (
		group *Group
		existed bool
	)
	if group, existed = self.bucket.groupMap[id]; !existed{
		group = InitGroup(id)

		self.bucket.rwMutex.Lock()
		self.bucket.groupMap[id] = group
		self.bucket.rwMutex.Unlock()
		return
	}

	group.Join(self)

	self.mutex.Lock()
	self.Groups[id] = group
	self.mutex.Unlock()
}

func (self *WSConn) joinBucket() {
	bucket := master.GetBucket(self)
	self.mutex.Lock()
	self.bucket = bucket
	self.mutex.Unlock()

	self.bucket.rwMutex.Lock()
	self.bucket.connMap[self.Id] = self
	self.bucket.rwMutex.Unlock()
}