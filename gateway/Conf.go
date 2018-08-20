package gateway

type Conf struct {
	inChanNum int
	outChanNum int
	heartbeatInterval int
	dispatchChanNum int
	bucketsNum uint64
	dispatchWorkerNum uint64
	msgInterval int
	mergerMsgNum int
	mergerTimeoutNum int
}

var (
	conf *Conf
)

func InitConf() {
	conf = &Conf{}
}