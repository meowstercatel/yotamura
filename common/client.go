package common

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Ws      *websocket.Conn                  `json:"-"`
	Name    string                           `json:"name"`
	Info    string                           `json:"info"`
	Actions map[string]func(message Message) `json:"-"`

	MessageChannel map[string]chan Message `json:"-"`
	mu             sync.RWMutex            `json:"-"`
}

func (c *Client) HandleMessages() {
	for {
		message := c.GetWsMessage()
		fmt.Println("message: ", message)

		if action, ok := c.Actions[message.Type]; ok {
			action(message)
		}
	}
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

	c.mu.Lock()
	c.MessageChannel[randString] = responseChannel
	c.mu.Unlock()

	message := <-responseChannel

	c.mu.Lock()
	delete(c.MessageChannel, randString)
	c.mu.Unlock()

	return message
}

func (c *Client) BroadcastWsMessage(message Message) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, ch := range c.MessageChannel {
		select {
		case ch <- message:
		default:
			// Channel full, message dropped
		}
	}
}
