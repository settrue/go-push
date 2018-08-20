package gateway

import (
	"github.com/settrue/go-push/common"
	"github.com/gorilla/websocket"
	"encoding/json"
)

func (self *WSConn)Handler() {
	var (
		msg *common.WSMsg
		err error
		bizReq *common.BizReq
	)
	for {
		if msg, err = self.Read();err != nil {
			break
		}

		if msg.Type != websocket.TextMessage {
			continue
		}

		if bizReq, err = msg.DataToBizReq(); err != nil {
			break
		}

		switch bizReq.Type {
		case common.PING:
			if err = self.handlerPing(bizReq); err != nil {
				break
			}
		case common.JOIN_GROUP:
			if err = self.handlerJoin(bizReq); err != nil {
				break
			}
		case common.LEAVE_GROUP:
			if err = self.handlerLeave(bizReq); err != nil {
				break
			}
		case common.OFFLINE:
			break
		}
	}

	self.Close()
	self.LeaveAll()
}

func (self *WSConn)handlerPing(b *common.BizReq) (err error) {
	var (
		bizResp *common.BizResp
		bt []byte
		wsMsg *common.WSMsg
	)
	bizResp = common.MkBizResp("PONG", bt)
	if bt, err = json.Marshal(bizResp); err != nil {
		return
	}
	wsMsg = common.MkWSMsg(websocket.TextMessage, bt)
	err = self.Write(wsMsg)
	return
}

func (self *WSConn)handlerJoin(b *common.BizReq) (err error) {
	var (
		groupId uint64
	)
	if groupId, err = b.GetGroup(); err != nil {
		return
	}
	self.JoinGroup(groupId)
	return
}

func (self *WSConn)handlerLeave(b *common.BizReq) (err error) {
	var (
		groupId uint64
	)
	if groupId, err = b.GetGroup(); err != nil {
		return
	}
	self.LeaveGroup(groupId)
	return
}