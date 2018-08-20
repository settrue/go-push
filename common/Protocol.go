package common

import "encoding/json"

type WSMsg struct{
	Type int
	Data []byte
}

func (self *WSMsg) DataToBizReq() (br *BizReq, err error){
	err = json.Unmarshal(self.Data, br)
	return
}

type BizReq struct{
	Type string
	Msg []byte
}

func (self *BizReq) GetGroup() (id uint64, err error){
	var br *BizGroup
	if err = json.Unmarshal(self.Msg, br); err != nil {
		return
	}
	id = br.GroupId
	return
}

type BizGroup struct {
	GroupId uint64 `json:"group_id"`
}

var (
	JOIN_GROUP = "JOIN GROUP"
	LEAVE_GROUP = "LEAVE GROUP"
	PING = "PING"
	OFFLINE = "OFFLINE"
)

func MkWSMsg(tp int, data []byte) *WSMsg{
	return &WSMsg{
		tp,
		data,
	}
}

type BizResp struct {
	Type string
	Msg []byte
}

func MkBizResp(tp string, data []byte) *BizResp {
	return &BizResp{
		tp,
		json.RawMessage(data),
	}
}