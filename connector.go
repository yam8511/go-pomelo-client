package client

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/yam8511/go-pomelo-client/codec"
	"github.com/yam8511/go-pomelo-client/message"
	"github.com/yam8511/go-pomelo-client/packet"
)

// Connector is a Pomelo client
type Connector struct {
	conn       net.Conn       // low-level connection
	codec      *codec.Decoder // decoder
	mid        uint           // message id
	muConn     sync.RWMutex
	connecting bool        // connection status
	die        chan byte   // connector close channel
	chSend     chan []byte // send queue

	// some packet data
	handshakeData    []byte // handshake data
	handshakeAckData []byte // handshake ack data
	heartbeatData    []byte // heartbeat data

	// events handler
	muEvents sync.RWMutex
	events   map[string]Callback

	// response handler
	muResponses sync.RWMutex
	responses   map[uint]Callback
}

// SetHandshake 設定握手協定
func (c *Connector) SetHandshake(handshake interface{}) error {
	data, err := json.Marshal(handshake)
	if err != nil {
		return err
	}

	c.handshakeData, err = codec.Encode(packet.Handshake, data)
	if err != nil {
		return err
	}

	return nil
}

// SetHandshakeAck 設定握手Ack協定
func (c *Connector) SetHandshakeAck(handshakeAck interface{}) error {
	var err error
	if handshakeAck == nil {
		c.heartbeatData, err = codec.Encode(packet.HandshakeAck, nil)
		if err != nil {
			return err
		}
		return nil
	}

	data, err := json.Marshal(handshakeAck)
	if err != nil {
		return err
	}

	c.handshakeAckData, err = codec.Encode(packet.HandshakeAck, data)
	if err != nil {
		return err
	}

	return nil
}

// SetHeartBeart 設定心跳包
func (c *Connector) SetHeartBeart(heartbeat interface{}) error {
	var err error
	if heartbeat == nil {
		c.heartbeatData, err = codec.Encode(packet.Heartbeat, nil)
		if err != nil {
			return err
		}
		return nil
	}
	data, err := json.Marshal(heartbeat)
	if err != nil {
		return err
	}

	c.heartbeatData, err = codec.Encode(packet.Heartbeat, data)
	if err != nil {
		return err
	}

	return nil
}

// Run 啟動連線
func (c *Connector) Run(addr string) error {
	if c.handshakeData == nil {
		return errors.New("handshake not defined")
	}

	if c.handshakeAckData == nil {
		c.SetHandshake(nil)
	}

	if c.heartbeatData == nil {
		c.SetHeartBeart(nil)
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	c.conn = conn

	go c.write()

	c.send(c.handshakeData)

	c.read()

	return nil
}

// Request send a request to server and register a callbck for the response
func (c *Connector) Request(route string, data []byte, callback Callback) error {
	msg := &message.Message{
		Type:  message.Request,
		Route: route,
		ID:    c.mid,
		Data:  data,
	}

	c.setResponseHandler(c.mid, callback)
	if err := c.sendMessage(msg); err != nil {
		c.setResponseHandler(c.mid, nil)
		return err
	}

	return nil
}

// Notify send a notification to server
func (c *Connector) Notify(route string, data []byte) error {
	msg := &message.Message{
		Type:  message.Notify,
		Route: route,
		Data:  data,
	}
	return c.sendMessage(msg)
}

// On add the callback for the event
func (c *Connector) On(event string, callback Callback) {
	c.muEvents.Lock()
	defer c.muEvents.Unlock()

	c.events[event] = callback
}

// Close close the connection, and shutdown the benchmark
func (c *Connector) Close() {
	if !c.connecting {
		return
	}
	// log.Println("連線關閉")
	c.conn.Close()
	c.die <- 1
	c.connecting = false
}

// IsClosed check the connection is closed
func (c *Connector) IsClosed() bool {
	return !c.connecting
}

func (c *Connector) eventHandler(event string) (Callback, bool) {
	c.muEvents.RLock()
	defer c.muEvents.RUnlock()

	cb, ok := c.events[event]
	return cb, ok
}

func (c *Connector) responseHandler(mid uint) (Callback, bool) {
	c.muResponses.RLock()
	defer c.muResponses.RUnlock()

	cb, ok := c.responses[mid]
	return cb, ok
}

func (c *Connector) setResponseHandler(mid uint, cb Callback) {
	c.muResponses.Lock()
	defer c.muResponses.Unlock()

	if cb == nil {
		delete(c.responses, mid)
	} else {
		c.responses[mid] = cb
	}
}

func (c *Connector) sendMessage(msg *message.Message) error {
	data, err := msg.Encode()
	if err != nil {
		return err
	}

	// log.Printf("資料 ---> %+v,\n整體 ---> %+v\n加密 ---> %+v\n", msg.Data, msg, data)

	payload, err := codec.Encode(packet.Data, data)
	if err != nil {
		return err
	}

	c.mid++
	c.send(payload)

	return nil
}

func (c *Connector) write() {
	for {
		select {
		case data := <-c.chSend:
			if _, err := c.conn.Write(data); err != nil {
				log.Println("傳送訊息失敗", err.Error())
				// c.Close()
			}

		case <-c.die:
			return
		}
	}
}

func (c *Connector) send(data []byte) {
	c.chSend <- data
}

func (c *Connector) read() {
	buf := make([]byte, 2048)

	for {
		time.Sleep(time.Second)
		if c.IsClosed() {
			return
		}
		n, err := c.conn.Read(buf)
		if err != nil {
			// log.Println("讀取資料失敗", err.Error())
			c.Close()
			return
			// continue
		}

		packets, err := c.codec.Decode(buf[:n])
		if err != nil {
			// log.Println("解碼資料失敗", err.Error())
			// c.Close()
			// return
			continue
		}

		for i := range packets {
			p := packets[i]
			// log.Println("讀取到資料包 --->", p)
			c.processPacket(p)
		}
	}
}

func (c *Connector) processPacket(p *packet.Packet) {
	switch p.Type {
	case packet.Handshake:
		var handShakeResponse struct {
			Code int `json:"code"`
			Sys  struct {
				Heartbeat int `json:"heartbeat"`
			} `json:"sys"`
		}

		err := json.Unmarshal(p.Data, &handShakeResponse)
		if err != nil {
			// log.Fatal("握手回傳進行解碼發生錯誤")
			c.Close()
			return
		}

		if handShakeResponse.Code == 200 {
			go func() {
				ticker := time.NewTicker(time.Second * time.Duration(handShakeResponse.Sys.Heartbeat))
				for range ticker.C {
					if c.IsClosed() {
						return
					}
					c.send(c.heartbeatData)
				}
			}()
			c.send(c.handshakeAckData)
		} else {
			// log.Fatal("握手回傳不是200狀態", string(p.Data))
			c.Close()
		}
	case packet.Data:
		msg, err := message.Decode(p.Data)
		if err != nil {
			log.Println(err.Error())
			return
		}
		c.processMessage(msg)

	case packet.Kick:
		// log.Fatal("Server 主動斷開連線通知 --->", p)
		c.Close()
	}
}

func (c *Connector) processMessage(msg *message.Message) {
	switch msg.Type {
	case message.Push:
		cb, ok := c.eventHandler(msg.Route)
		if !ok {
			log.Println("event handler not found", msg.Route)
			return
		}

		cb(msg.Data)

	case message.Response:
		cb, ok := c.responseHandler(msg.ID)
		if !ok {
			log.Println("response handler not found", msg.ID)
			return
		}

		cb(msg.Data)
		c.setResponseHandler(msg.ID, nil)
	}
}
