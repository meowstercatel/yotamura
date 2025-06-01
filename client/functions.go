package main

import (
	"fmt"
	"os"
	"yotamura/common"
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

	c.Actions["FileData"] = func(message common.Message) {
		fmt.Println("file")
		var content common.FileData
		common.DecodeData(message.Data, &content)
		var files []common.File
		osFiles, err := os.ReadDir(content.Path)
		if err != nil {
			fmt.Println("can't read dir ", err)
		}
		for _, file := range osFiles {
			files = append(files, common.File{Name: file.Name(), IsDirectory: file.Type().IsDir()})
		}
		fmt.Println(files)
		c.SendJsonMessage(common.CreateMessage(common.FileData{Path: content.Path, Files: files}))
	}
	c.Actions["ReadFileData"] = func(message common.Message) {
		fmt.Println("read")
		var content common.ReadFileData
		common.DecodeData(message.Data, &content)
		output, err := os.ReadFile(content.Path)
		if err != nil {
			//handle err
		}
		c.SendJsonMessage(common.CreateMessage(common.ReadFileData{Path: content.Path, Content: output}))
	}
}
