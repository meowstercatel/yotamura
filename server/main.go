package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"yotamura/common"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

type Client struct {
	*common.Client
}

func (c *Client) HandleMessages() {
	for {
		message := c.GetWsMessage()
		fmt.Println(message)

		switch message.Type {
		case "CommandData":
			fmt.Println("command")
			var content common.CommandData
			common.DecodeData(message.Data, &content)
			fmt.Println(content)
		case "StatsData":
			fmt.Println("stats")
			var content common.StatsData
			common.DecodeData(message.Data, &content)
			c.Name = content.Name
		default:
			fmt.Printf("Unknown message type: %s\n", message.Type)
		}
	}
}

func (c *Client) SendCommand(command string, waitForResponse bool) (common.CommandData, error) {
	err := c.SendJsonMessage(common.CreateMessage(common.CommandData{Command: command}))
	fmt.Println("sent message")
	if err != nil {
		return common.CommandData{}, err
	}

	if !waitForResponse {
		return common.CommandData{}, nil
	}
	for {
		fmt.Println("waiting for message")
		response := c.GetWsMessage()
		var content common.CommandData
		err := common.DecodeData(response.Data, &content)
		if err != nil {
			//the first response doesn't always have to be the command result
			continue
		}
		return content, nil
	}
}

var clients []Client

func RemoveIndex(s []Client, index int) []Client {
	return append(s[:index], s[index+1:]...)
}

func handle(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade err:", err)
		return
	}
	defer c.Close()

	client := Client{
		Client: &common.Client{
			Ws:             c,
			MessageChannel: make(map[string]chan common.Message),
		},
	}

	clients = append(clients, client)

	go client.HandleMessages()
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			if closeError, ok := err.(*websocket.CloseError); ok {
				log.Printf("websocket closed! code: %d", closeError.Code)
			} else {
				log.Println("read err:", err)
			}
			for i, v := range clients {
				if client.Name == v.Name {
					clients = RemoveIndex(clients, i)
				}
			}
			break
		}

		message := common.Message{}
		err = json.Unmarshal(msg, &message)
		if err != nil {
			fmt.Println("failed to convert message to struct")
		}
		client.BroadcastWsMessage(message)
	}
}

func send(w http.ResponseWriter, r *http.Request) {
	requestJson, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("failed to read request!", err)
		w.WriteHeader(400)
		w.Write([]byte("failed to read request!"))
		return
	}
	var requestMessage common.RequestMessage
	err = json.Unmarshal(requestJson, &requestMessage)
	if err != nil {
		fmt.Println("received bad data!", err)
		w.WriteHeader(400)
		w.Write([]byte("received bad data!"))
		return
	}

	switch requestMessage.Message.Type {
	case "CommandData":
		fmt.Println("command")
		var content common.CommandData
		common.DecodeData(requestMessage.Message.Data, &content)
		fmt.Println(content)
		client := clients[requestMessage.SendTo]
		response, err := client.SendCommand(content.Command, requestMessage.WaitForResponse)
		if err != nil {
			fmt.Println("sending command failed!", err)
			w.WriteHeader(500)
			w.Write(fmt.Appendf(nil, "sending command failed!, %v", err))
			return
		}
		responseJson, _ := json.Marshal(common.CreateMessage(response)) //no error = no problem
		w.Write(responseJson)
	default:
		fmt.Printf("Unknown message type: %s\n", requestMessage.Message.Type)
	}
}

func returnClients(w http.ResponseWriter, r *http.Request) {
	jsonClients, err := json.Marshal(clients)
	if err != nil {
		fmt.Println("error converting clients to json")
	}
	w.Write(jsonClients)
}

func main() {
	flag.Parse()
	http.HandleFunc("/ws", handle)
	http.HandleFunc("/send", send)
	http.HandleFunc("/clients", returnClients)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
