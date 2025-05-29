package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"yotamura/common"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

type Client struct {
	Ws   *websocket.Conn `json:"-"`
	Name string          `json:"name"`
	Info string          `json:"info"`

	Message chan common.Message `json:"-"`
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

func (c *Client) SendMessage(data []byte) error {
	return c.Ws.WriteMessage(1, data)
}

func (c *Client) HandleMessages() {
	for {
		message := <-c.Message
		fmt.Println(message)

		switch message.Data.(type) {
		case common.CommandData:
			fmt.Println("command")
			data := message.Data.(*common.CommandData)
			fmt.Println(data)
		default:

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

	client := Client{Ws: c}
	clients = append(clients, client)

	go client.HandleMessages()
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			if closeError, ok := err.(*websocket.CloseError); ok {
				log.Printf("websocket closed! code: %d", closeError.Code)

				clients = RemoveIndex(clients, slices.Index(clients, client))
			} else {
				log.Println("read err:", err)
			}
			break
		}

		message := common.Message{}
		err = json.Unmarshal(msg, &message)
		if err != nil {
			fmt.Println("error when converting message to json")
		}
		client.Message <- message
		// client.handleMessage(mt, message)
		// log.Printf("recv: %s", message)
		// err = c.WriteMessage(mt, message)
		// if err != nil {
		// 	log.Println("write err:", err)
		// 	break
		// }
	}
}

func send(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	//shitty implementation for now
	command := query.Get("command")
	clientId, _ := strconv.ParseInt(query.Get("clientId"), 10, 32)

	fmt.Println("got /send request")

	message := common.Message{Data: common.CommandData{Command: command}}
	jsonMessage, err := json.Marshal(message)
	if err != nil { //this should never fail tho
		fmt.Println("error while trying to marshal message")
	}

	client := clients[clientId]
	client.SendMessage(jsonMessage)

	for {
		response := <-client.Message
		//the first response doesn't always have to be the command result
		if command, ok := response.Data.(*common.CommandData); ok {
			jsonResponse, _ := json.Marshal(command)
			w.Write(jsonResponse)
			return
		}
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
