package common

import "qiniupkg.com/x/errors.v7"

var (
	SOCKET_HAVE_CLOSE = errors.New("The websocket have closed")
	OUT_CHAN_IS_FULL = errors.New("Out chan is full")

	CONN_HAVE_EXISTS_GROUP = errors.New("The WSConn have exists in the group")
	CONN_HAVE_NOT_EXISTS_GROUP = errors.New("The WSConn have not exists in the group")
)