package common

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
)

func GetType(variable any) string {
	typeName := strings.Split(fmt.Sprintf("%T", variable), ".")
	return typeName[len(typeName)-1]
}

func CreateMessage(data any) Message {
	return Message{Type: GetType(data), Data: data}
}

func DecodeData(input any, output any) error {
	return mapstructure.Decode(input, output)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func CopyFile(src, dst string) (err error) {
	// Open the source file.
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	// Create the destination file.
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	// Copy the contents from the source to the destination.
	_, err = io.Copy(out, in)
	return
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		// Handle other potential errors here if needed
		fmt.Println("Error checking path:", err)
		return false
	}
	return true
}
