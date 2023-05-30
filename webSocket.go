package gosocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"time"
)

// tttt
type SocketConn struct {
	SocketChannel    chan string
	SocketConnection *websocket.Conn
}

var SocketConnections = make(map[int]*SocketConn, 3)
var running = false

func CreateSocketConn(host string, port int16, appId *int, funcs map[string]interface{}) bool {
	u := url.URL{Scheme: "wss", Host: fmt.Sprintf("%s:%d", host, port), Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return false
	}

	SocketConnections[*appId] = &SocketConn{
		SocketChannel:    make(chan string),
		SocketConnection: c,
	}

	s := SocketConnections[*appId]
	go s.runSocket(*appId, funcs)
	return true
}

func (s *SocketConn) waitForMessage(c chan string) {
	for running {
		_, message, err := s.SocketConnection.ReadMessage()
		if err != nil {
			log.Println("Could not read message from socket", err)
			return
		}
		println(message)
		c <- string(message)
	}
}

func (s *SocketConn) runSocket(appId int, funcs map[string]interface{}) {
	t := time.Now().Unix()
	msgChan := make(chan string)
	running = true
	go s.waitForMessage(msgChan)
	for {
		select {
		case v := <-s.SocketChannel:
			if v == "close" {
				err := s.SocketConnection.Close()
				if err != nil {
					return
				}
				close(s.SocketChannel)
				running = false
				break
			}
			err := s.SocketConnection.WriteMessage(websocket.TextMessage, []byte(v))
			if err != nil {
			}
		case message := <-msgChan:
			fn, ok := funcs[message].(func(string) (bool, error))
			if ok {
				_, err := fn(message)
				if err != nil {
				}
			}
		default:
			if time.Now().Unix()-t >= 60 {
				err := s.SocketConnection.WriteMessage(websocket.TextMessage, []byte("KEAL"))
				if err != nil {
				}
				t = time.Now().UnixNano()
			}
		}

		time.Sleep(200 * time.Millisecond)
	}
}
