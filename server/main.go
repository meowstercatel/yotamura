package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"yotamura/common"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

type Client struct {
	*common.Client
}

// func (c *Client) handleMessage(mt int, msg []byte) {
// 	c.Mt = mt

// 	message := Message{}
// 	err := json.Unmarshal(msg, &message)
// 	if err != nil {
// 		fmt.Println("error parsing message")
// 	}
// 	switch message.Data.(type) {
// 	case CommandData:
// 		fmt.Println("command")
// 		data := message.Data.(*CommandData)
// 		fmt.Println(data)
// 	case string:

// 	default:

// 	}
// }

func DecodeData(input any, output any) error {
	return mapstructure.Decode(input, output)
}

func (c *Client) HandleMessages() {
	for {
		// message := <-c.Message
		message := c.GetWsMessage()
		fmt.Println(message)

		switch message.Type {
		case "CommandData":
			fmt.Println("command")
			var content common.CommandData
			DecodeData(message.Data, &content)
			fmt.Println(content)
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
			Message:        make(chan common.Message),
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
		// client.Message <- message
		client.BroadcastWsMessage(message)
	}
}

func send(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	//shitty implementation for now
	command := query.Get("command")
	clientId, _ := strconv.ParseInt(query.Get("clientId"), 10, 32)

	fmt.Println("got /send request")

	if len(clients) <= int(clientId) {
		w.Write([]byte("bad client id!"))
		return
	}

	client := clients[clientId]
	err := client.SendJsonMessage(common.CreateMessage(common.CommandData{Command: command}))
	fmt.Println("sent message")
	if err != nil {
		fmt.Println("error when sending message", err)
	}

	for {
		// response := <-client.Message
		fmt.Println("waiting for message")
		response := client.GetWsMessage()
		fmt.Println("/send: response: ", response)
		//the first response doesn't always have to be the command result
		var content common.CommandData
		err := DecodeData(response.Data, &content)
		if err != nil {
			fmt.Println("/send ERROR DECODING MESSAGE", err)
			continue
		}
		jsonResponse, _ := json.Marshal(content)
		w.Write(jsonResponse)
		return
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
