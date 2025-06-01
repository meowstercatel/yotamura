package main

import (
	"fmt"
	"yotamura/common"
)

func (c *Client) initializeHandlers() {
	c.Actions = make(map[string]func(message common.Message))
	c.Actions["StatsData"] = func(message common.Message) {
		fmt.Println("stats")
		var content common.StatsData
		common.DecodeData(message.Data, &content)
		c.Name = content.Name
	}
}
