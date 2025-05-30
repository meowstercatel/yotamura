package common

import "github.com/gorilla/websocket"

type Client struct {
	Ws   *websocket.Conn `json:"-"`
	Name string          `json:"name"`
	Info string          `json:"info"`

	Message chan Message `json:"-"`
}

func (c *Client) SendMessage(data []byte) error {
	return c.Ws.WriteMessage(1, data)
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
