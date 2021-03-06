package controllers

import (
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"goApiServer/lang"
	"goApiServer/ws"
	"net/http"
	"time"
)

type API struct {
	beego.Controller
}

// 函数被执行后，就可以通过这个全局变量，在全局获取语言数据
func (api *API) GetLang() {
	// 设置语言
	var userLang string
	// 语言处理
	lg := api.Ctx.Request.Header.Get("Accept-Language")
	if lg != "" {
		// 语言设置不为空，就使用用户定义的语言
		userLang = lg
	} else {
		// 否则就使用默认语言
		userLang = beego.AppConfig.String("lang")
	}

	// 重新设置语言参数
	lang.SetLang(userLang)

	// IP处理
	ip := api.Ctx.Request.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = api.Ctx.Request.Header.Get("X-real-ip")
	}
	if ip == "" {
		ip = api.Ctx.Input.IP()
	}
	if ip != "" {
		// 用户的IP地址:
		beego.Trace("用户的IP地址:", ip)
	} else {
		// 无法获取用户的IP地址:(
		beego.Trace("无法获取用户的IP地址!")
	}
}

type WebSocketController struct {
	beego.Controller
}

var (
	// Configure the upgrader
	upgrader = websocket.Upgrader{
		// 读取存储空间大小
		ReadBufferSize: 1024,
		// 写入存储空间大小
		WriteBufferSize: 1024,
		// 允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func (c *WebSocketController) SayHi() {
	// 定义变量
	var (
		wsConn *websocket.Conn
		err    error
		// data []byte
		conn *ws.Connection
		data []byte
	)

	// 完成http应答，升级为 websocket 连接
	if wsConn, err = upgrader.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil); err != nil {
		return // 获取连接失败直接返回
	}

	// 开始接收处理 websocket 请求
	if conn, err = ws.InitConnection(wsConn); err != nil {
		goto ERR
	}

	// 心跳
	go func() {
		var (
			err error
		)
		for {
			// 每隔一秒发送一次心跳，内容要转化为字节码，不能直接使用字符串。
			if err = conn.WriteMessage([]byte("heartbeat")); err != nil {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		// 调用方法查看 channel 已经接收的数据
		if data, err = conn.ReadMessage(); err != nil {
			goto ERR
		}
		beego.Debug("收到一条消息:", string(data))

		// 发送数据，处理协程收到 channel 消息后自动发出。
		if err = conn.WriteMessage(data); err != nil {
			goto ERR
		}
	}

ERR:
	// 关闭连接
	conn.Close()
	// 随便返回一条信息，否则 beego 会报找不到html模板的错误!
	c.Ctx.WriteString("")
}
