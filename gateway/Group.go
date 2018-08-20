package gateway

import (
	"sync"
	"github.com/settrue/go-push/common"
)

type Group struct{
	Id uint64
	rwMutex sync.RWMutex
	ConnMap map[uint64]*WSConn
}

func InitGroup(id uint64) *Group {
	return &Group{
		Id:id,
		ConnMap:make(map[uint64]*WSConn),
	}
}

func (self *Group)Join(conn *WSConn) (err error) {
	if _, exists := self.ConnMap[conn.Id]; exists {
		err = common.CONN_HAVE_EXISTS_GROUP
		return
	}

	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	self.ConnMap[conn.Id] = conn
	return
}

func (self *Group)Leave(conn *WSConn) (err error) {
	if _, exists := self.ConnMap[conn.Id]; !exists {
		err = common.CONN_HAVE_NOT_EXISTS_GROUP
		return
	}

	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	delete(self.ConnMap, conn.Id)
	return
}

func (self *Group)Count() int {
	self.rwMutex.RLock()
	defer self.rwMutex.RUnlock()

	return len(self.ConnMap)
}

func (self *Group)Push(wsMsg *common.WSMsg) {
	self.rwMutex.RLock()
	defer self.rwMutex.RUnlock()

	var (
		conn *WSConn
	)
	for _, conn = range self.ConnMap {
		conn.Write(wsMsg)
	}
}