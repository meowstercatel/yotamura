package common

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Client struct {
	Ws   *websocket.Conn `json:"-"`
	Name string          `json:"name"`
	Info string          `json:"info"`

	Message        chan Message            `json:"-"`
	MessageChannel map[string]chan Message `json:"-"`
}

func (c *Client) SendMessage(data []byte) error {
	return c.Ws.WriteMessage(1, data)
}

func (c *Client) SendJsonMessage(data any) error {
	return c.Ws.WriteJSON(data)
}

func (c *Client) GetWsMessage() Message {
	responseChannel := make(chan Message, 100)
	randString := RandString(5)

	c.MessageChannel[randString] = responseChannel

	message := <-responseChannel // This will block until BroadcastWsMessage sends
	fmt.Println("GOT MESSAGE: ", message)

	delete(c.MessageChannel, randString)
	return message
}

func (c *Client) BroadcastWsMessage(message Message) {
	for _, ch := range c.MessageChannel {
		select {
		case ch <- message:
		default:
			// Channel full, message dropped (or you could make it blocking)
		}
	}
}

// func (c *Client) HandleMessages() {
// 	for {
// 		message := <-c.Message
// 		fmt.Println(message)

// 		switch message.Data.(type) {
// 		case common.CommandData:
// 			fmt.Println("command")
// 			data := message.Data.(*common.CommandData)
// 			fmt.Println(data)
// 		default:

// 		}
// 	}
// }
