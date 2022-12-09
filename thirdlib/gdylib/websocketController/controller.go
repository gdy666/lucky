package websocketcontroller

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	disconnected uint32 = iota
	connecting
	reconnecting
	connected
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 9 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 1500 * time.Millisecond
	// Maximum message size allowed from peer.
	maxMessageSize = 4096 * 4

	handshakeTimeout = 10 * time.Second

	readTimeout = 9 * time.Second
)

type Controller struct {
	Logs     *logrus.Logger
	mu       sync.Mutex
	wait     sync.WaitGroup
	initOnce sync.Once //一次

	status      uint32
	readTimeout time.Duration
	writeWait   time.Duration //自定义写超时参数
	pongWait    time.Duration //自定义Pong超时参数
	pingPeriod  time.Duration //自定义发送ping时间间隔
	sendChan    chan []byte

	ReceiveMessageCallback       func(c *Controller, messageBytes []byte) //接收消息回调函数
	ctx                          context.Context
	ctxCancelFunc                context.CancelFunc
	sendMessageEncryptionFunc    func(messageBytes []byte) ([]byte, error) //发送消息加密函数
	receiveMessageDecryptionFunc func(messageBytes []byte) ([]byte, error) //接收消息解密函数

	//---------- 客户端特有
	serverUrl                  string
	scureSkipVerify            bool
	handshakeTimeout           time.Duration
	disconnectSign             uint32 //主动断开连接标记,不自动重连
	connectRetry               bool   //重连
	connectRetryInterval       time.Duration
	connectRetryCount          int //0表示无限次
	ClientReadyCallback        func(c *Controller)
	ClientDisconnectedCallback func(c *Controller)

	remoteAddr string
	extDataMap sync.Map
}

func (c *Controller) SetSendMessageEncryptionFunc(f func(messageBytes []byte) ([]byte, error)) {
	c.sendMessageEncryptionFunc = f
}

func (c *Controller) SetReceiveMessageDecryptionFunc(f func(messageBytes []byte) ([]byte, error)) {
	c.receiveMessageDecryptionFunc = f
}

func (c *Controller) StoreExtData(key any, data any) {
	c.extDataMap.Store(key, data)

}

func (c *Controller) GetExtData(key any) (any, bool) {
	return c.extDataMap.Load(key)
}

func (c *Controller) disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()
	status := atomic.LoadUint32(&c.status)
	if status == disconnected {
		return
	}

	c.wait.Wait()

	if c.ctxCancelFunc != nil {
		c.ctxCancelFunc()
		c.ctxCancelFunc = nil
	}

	if c.ClientDisconnectedCallback != nil {
		c.ClientDisconnectedCallback(c)
	}
	atomic.StoreUint32(&c.status, disconnected)

	disconnectSign := atomic.LoadUint32(&c.disconnectSign)

	//	if c.connectRetry && !c.disconnectSign {
	if c.connectRetry && disconnectSign == 0 {
		go func() {
			<-time.After(c.connectRetryInterval)
			c.Connect()
		}()
	}

}

func (c *Controller) SetServerURL(url string) {
	c.serverUrl = url

}

func (c *Controller) ScureSkipVerify(b bool) {
	c.scureSkipVerify = b
}

func (c *Controller) SetConnectRetry(b bool) {
	c.connectRetry = b
}

func (c *Controller) SetConnectRetryInterval(t time.Duration) {
	c.connectRetryInterval = t
}

func (c *Controller) SetConnectRetryCount(count int) {
	c.connectRetryCount = count
}

func (c *Controller) SetReadDeadline(d time.Duration) {
	c.readTimeout = d
}

func (c *Controller) init() {
	c.mu.Lock()
	defer c.mu.Unlock()
	status := atomic.LoadUint32(&c.status)
	if status == connected {
		return
	}

	if c.writeWait <= 0 {
		c.writeWait = writeWait
	}

	if c.pongWait <= 0 {
		c.pongWait = pongWait
	}

	if c.pingPeriod <= 0 {
		c.pingPeriod = pingPeriod
	}

	if c.handshakeTimeout <= 0 {
		c.handshakeTimeout = handshakeTimeout
	}

	if c.readTimeout <= 0 {
		c.readTimeout = readTimeout
	}

	//c.disconnectSign = false
	atomic.StoreUint32(&c.disconnectSign, 0)

	c.initOnce.Do(func() {
		c.sendChan = make(chan []byte, 1024)
	})

}

func (c *Controller) Disconnect() {

	//c.disconnectSign = true
	atomic.StoreUint32(&c.disconnectSign, 1)
	c.mu.Lock()
	defer c.mu.Unlock()
	status := atomic.LoadUint32(&c.status)
	if status != connected {
		return
	}
	c.Logs.Infof("[%s]Disconnect", c.serverUrl)
	if c.ctxCancelFunc != nil {
		c.ctxCancelFunc()
	}

}

func (c *Controller) SendMessage(messageBytes []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	status := atomic.LoadUint32(&c.status)
	if status != connected {
		c.Logs.Warn("websocketclient 未连接,消息[%s]发送失败", string(messageBytes))
		return
	}

	if c.sendMessageEncryptionFunc != nil {
		enMsgBytes, err := c.sendMessageEncryptionFunc(messageBytes)
		if err != nil {
			c.Logs.Error("WebSocket客户端 自定义发送消息加密函数加密出错:%s", err.Error())
			return
		}
		messageBytes = enMsgBytes
	}

	select {
	case c.sendChan <- messageBytes:
	default:
	}
}

func (c *Controller) writePump(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(c.pingPeriod)
	c.wait.Add(1)

	defer func() {

		conn.Close()
		c.wait.Done()
		ticker.Stop()
		c.disconnect()
		//c.Logs.Printf("writePump return\n")
		recoverErr := recover()
		if recoverErr == nil {
			return
		}
		c.Logs.Errorf("webscoket controller writePump panic:\n%v", recoverErr)
		c.Logs.Errorf("%s", debug.Stack())
	}()
	//ccontext, _ := context.WithCancel(ctx)
	for {

		select {
		case messageBytes, ok := <-c.sendChan: //需要发送的消息

			conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			if len(messageBytes) == 0 {
				continue
			}

			w.Write(messageBytes)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		// case <-ccontext.Done():
		// 	{
		// 		return
		// 	}
		case <-ctx.Done():
			{
				//fmt.Printf("ctx.Done():666\n")
				return
			}
		}
	}

}

func (c *Controller) readPump(cancelFunc context.CancelFunc, conn *websocket.Conn) {
	defer func() {
		c.wait.Done()
		cancelFunc()
		//c.Logs.Printf("readPump return ")
		recoverErr := recover()
		if recoverErr == nil {
			return
		}
		c.Logs.Errorf("webscoket controller readPump panic:\n%v", recoverErr)
		c.Logs.Errorf("%s", debug.Stack())
	}()

	c.wait.Add(1)
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(c.pongWait))
		return nil
	})

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Logs.Error("websocket.IsUnexpectedCloseError : %s", err.Error())
			}
			break
		}

		messageSize := len(messageBytes)
		if messageSize <= 1 { //忽略心跳消息
			continue
		}

		if c.receiveMessageDecryptionFunc != nil {
			deMsgBytes, err := c.receiveMessageDecryptionFunc(messageBytes)
			if err != nil {
				c.Logs.Error("Websocket controller 消息接收消息解密出错:%s", err.Error())
				continue
			}
			messageBytes = deMsgBytes
		}

		if c.ReceiveMessageCallback != nil {
			go c.ReceiveMessageCallback(c, messageBytes)
		}
	}

}

func (c *Controller) Connect() error {
	c.init()
	c.mu.Lock()
	defer c.mu.Unlock()

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: c.handshakeTimeout,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: c.scureSkipVerify}}

	atomic.StoreUint32(&c.status, connecting)
	retryCount := 0
RETRYCONN:
	connect, _, err := dialer.Dial(c.serverUrl, nil)
	if err != nil {
		disconnectSign := atomic.LoadUint32(&c.disconnectSign)
		//	if c.connectRetry && (c.connectRetryCount <= 0 || retryCount < c.connectRetryCount) && !c.disconnectSign {
		if c.connectRetry && (c.connectRetryCount <= 0 || retryCount < c.connectRetryCount) && disconnectSign == 0 {
			<-time.After(c.connectRetryInterval)
			retryCount++
			c.Logs.Errorf("[%d]Connect error:%s", disconnectSign, err.Error())
			c.Logs.Infof("[%s]重新连接...%d\n", c.serverUrl, retryCount)
			goto RETRYCONN
		}
		atomic.StoreUint32(&c.status, disconnected)
		return fmt.Errorf("Connect DefaultDialer error:%s", err.Error())
	}
	//c.conn = connect

	if c.ctxCancelFunc != nil {
		c.ctxCancelFunc()
	}

	c.monitor(connect)

	return nil
}

// ConnectReady Websocket 服务器端使用
func (c *Controller) ConnectReady(connect *websocket.Conn) {
	c.init()
	c.mu.Lock()
	defer c.mu.Unlock()

	c.remoteAddr = connect.RemoteAddr().String()
	c.monitor(connect)
}

func (c *Controller) GetRemoteAddr() string {
	return c.remoteAddr
}

func (c *Controller) monitor(connect *websocket.Conn) {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())
	c.ctx = ctx
	c.ctxCancelFunc = ctxCancelFunc
	atomic.StoreUint32(&c.status, connected)

	go c.writePump(ctx, connect)
	go c.readPump(ctxCancelFunc, connect)

	if c.ClientReadyCallback != nil {
		go c.ClientReadyCallback(c)
	}
}
