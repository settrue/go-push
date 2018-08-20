package common

type Conn interface {
	readLoop()
	writeLoop()
	Read() (*WSMsg, error)
	Write(*WSMsg) error
	Close()
	IsAlive() bool
	KeepAlive()
}