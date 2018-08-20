package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sync/atomic"
	"strconv"
)

var (
	connNowId uint64
)

func InitService(){
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	var (
		upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	)

	router.Any("/ws", func(c *gin.Context) {
		var (
			err error
			ws *websocket.Conn
			connId uint64
			wsConn *WSConn
		)

		if ws, err = upgrader.Upgrade(c.Writer, c.Request, nil); err != nil {
			return
		}

		connId = atomic.AddUint64(&connNowId, 1)
		wsConn = InitConn(connId, ws)

		wsConn.Handler()
	})

	router.POST("/ws/push_group", func(c *gin.Context) {
		var (
			groupId uint64
			tmpStr  string
			tmpInt  int
			err     error
			msg     string
		)
		tmpStr = c.PostForm("group_id")
		if tmpInt, err = strconv.Atoi(tmpStr); err != nil {
			c.JSON(400, gin.H{
				"status": "Bed Request!",
			})
			return
		}
		groupId = uint64(tmpInt)
		msg = c.PostForm("msg")
		merger.Write(groupId, []byte(msg))
	})

	router.Run(":7777")
}