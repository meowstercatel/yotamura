package main

import (
	"fmt"
	"sync"
	"yotamura/common"
)

type Client struct {
	*common.Client
	CommandReceiveChannel map[string]chan []byte `json:"-"`
	CommandSendChannel    map[string]chan []byte `json:"-"`

	mu sync.RWMutex `json:"-"`
}

func (c *Client) NewCommand(channelId string) {
	channel := make(chan []byte, 100)
	// randString := common.RandString(5)

	c.mu.Lock()
	c.CommandReceiveChannel[channelId] = channel
	c.CommandSendChannel[channelId] = channel
	c.mu.Unlock()
	return
}

func (c *Client) RemoveCommand(id string) {
	c.mu.Lock()
	delete(c.CommandReceiveChannel, id)
	delete(c.CommandSendChannel, id)
	c.mu.Unlock()
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
