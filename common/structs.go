package common

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type RequestMessage struct {
	Message Message `json:"message"`
	SendTo  int     `json:"send_to"` //client to send the message to
}

type File struct {
	Name        string `json:"name"`
	IsDirectory bool   `json:"is_directory"`
}

type CommandData struct {
	Command       string `json:"command"`
	Output        string `json:"output"`
	WaitForOutput bool   `json:"wait_for_output"`
}
type StatsData struct {
	Name string `json:"name"`
}
type ErrorData struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}
type DirectoryData struct {
	Path  string `json:"path"`
	Files []File `json:"files"`
}
type ReadFileData struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}
type ScreenshotData struct {
	Screenshot []byte `json:"screenshot"`
}
