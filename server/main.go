package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"yotamura/common"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

var clients []*Client

func RemoveIndex(s []*Client, index int) []*Client {
	return append(s[:index], s[index+1:]...)
}

func handle(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade err:", err)
		return
	}
	defer c.Close()

	client := &Client{
		Client: &common.Client{
			Ws:             c,
			MessageChannel: make(map[string]chan common.Message),
		},
		CommandReceiveChannel: make(map[string]chan []byte),
		CommandSendChannel:    make(map[string]chan []byte),
	}
	go client.initializeHandlers()
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

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		next(w, r)
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

	client := clients[requestMessage.SendTo]

	client.SendJsonMessage(requestMessage.Message)
	for {
		response := client.GetWsMessage()
		if requestMessage.Message.Type == "ErrorData" {
			var clientErr common.ErrorData
			common.DecodeData(response.Data, &clientErr)
			if clientErr.Type == requestMessage.Message.Type {
				responseJson, _ := json.Marshal(common.CreateMessage(response))
				w.Write(responseJson)
			}
		}
		if requestMessage.Message.Type == response.Type {
			responseJson, _ := json.Marshal(common.CreateMessage(response))
			w.Write(responseJson)
			break
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

// it'd make sense to send raw data to this socket
// instead of a message struct
func CommandWs(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade err:", err)
		return
	}
	defer c.Close()

	var client *Client
	hostname := r.Header.Get("X-Client-Hostname")
	channelId := r.Header.Get("X-Command")

	for _, cl := range clients {
		if cl.Name == hostname {
			client = cl
		}
	}

	client.NewCommand(channelId)

	commandSendChannel := client.CommandSendChannel[channelId]
	go func() {
		data := <-commandSendChannel
		c.WriteMessage(int(ws.OpText), data)
	}()
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			if closeError, ok := err.(*websocket.CloseError); ok {
				log.Printf("websocket closed! code: %d", closeError.Code)
			} else {
				log.Println("read err:", err)
			}

			client.RemoveCommand(channelId)
			break
		}
		client.CommandReceiveChannel[channelId] <- msg
	}
}

func SendCommandWs(w http.ResponseWriter, r *http.Request) {
	//client sends a request to this endpoint
	//this sends a request to a given client(pc)
	//then the client connects with /commandWs
	//*dark magic* and now in THIS func the client(pc) communicates with the client(web).

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		// handle error
	}
	defer conn.Close()

	clientId, err := strconv.Atoi(r.URL.Query().Get("client"))
	if err != nil {
		log.Println("failed to convert client id to int")
	}

	command := r.URL.Query().Get("command")
	client := clients[clientId]

	client.SendJsonMessage(common.CreateMessage(common.CommandData{Command: command, Websocket: true}))

	//honestly this logic is only useful when the client:
	//has a 10mhz cpu that can't handle anything
	//or when the client is bombarded with commands asynchronously
	for {
		msg := client.GetWsMessage()
		var commandData common.CommandData
		if msg.Type == "CommandData" {
			common.DecodeData(msg.Data, &commandData)
			if commandData.Websocket {
				break
			}
		}
	}

	//we can assume that the channel exists bc of the logic above
	commandReceiveChannel := client.CommandReceiveChannel[command]
	commandSendChannel := client.CommandSendChannel[command]

	go func() {
		data := <-commandReceiveChannel
		err = wsutil.WriteServerMessage(conn, ws.OpText, data)
		if err != nil {
			// handle error
		}
	}()

	for {
		msg, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			// handle error
		}
		commandSendChannel <- msg
	}
}

func main() {
	flag.Parse()
	http.HandleFunc("/ws", handle)
	http.HandleFunc("/commandWs", CommandWs)
	http.HandleFunc("/sendCommandWs", SendCommandWs)
	http.HandleFunc("/send", CORS(send))
	http.HandleFunc("/clients", CORS(returnClients))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
