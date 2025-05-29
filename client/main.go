package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yotamura/common"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

type Client struct {
	Ws   *websocket.Conn `json:"-"`
	Name string          `json:"name"`
	Info string          `json:"info"`

	Message chan common.Message `json:"-"`
}

var client Client

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	fmt.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("dial:", err)
		fmt.Println("reconnecting in 10s")
		time.Sleep(10 * time.Second)
		go main()
	}
	defer c.Close()

	done := make(chan struct{})

	client = Client{Ws: c}
	go func() {
		defer close(done)
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				// fmt.Printf("%t", err)
				fmt.Println("reconnecting in 10s")
				time.Sleep(10 * time.Second)
				go main()
				return
			}

			message := common.Message{}
			err = json.Unmarshal(msg, &message)
			if err != nil {
				fmt.Println("error when converting message to json")
			}
			client.Message <- message
			fmt.Printf("recv: %s", msg)
		}
	}()

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)

	<-channel
	os.Exit(0)

	// ticker := time.NewTicker(time.Second)
	// defer ticker.Stop()

	// for {
	// 	select {
	// 	case <-done:
	// 		return
	// 	case t := <-ticker.C:
	// 		fmt.Println(t)
	// 		 err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
	// 		 if err != nil {
	// 		 	fmt.Println("write:", err)
	// 			return
	// 		 }
	// 	case <-interrupt:
	// 		fmt.Println("interrupt")

	// 		// Cleanly close the connection by sending a close message and then
	// 		// waiting (with timeout) for the server to close the connection.
	// 		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// 		if err != nil {
	// 			fmt.Println("write close:", err)
	// 			return
	// 		}
	// 		select {
	// 		case <-done:
	// 		case <-time.After(time.Second):
	// 		}
	// 		return
	// 	}
	// }
}
