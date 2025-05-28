package main

import (
	"flag"
	"log"
	"net/http"
	"slices"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

type Client struct {
	Ws   *websocket.Conn
	Name string
	Info string
}

type Message struct {
	Operation int
	Data      interface{}
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
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			if closeError, ok := err.(*websocket.CloseError); ok {
				log.Printf("websocket closed! code: %d", closeError.Code)

				clients = RemoveIndex(clients, slices.Index(clients, client))
			} else {
				log.Println("read err:", err)
			}
			break

		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write err:", err)
			break
		}
	}
}

func main() {
	flag.Parse()
	http.HandleFunc("/ws", handle)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
