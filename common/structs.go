package common

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type RequestMessage struct {
	Message         Message `json:"message"`
	SendTo          int     `json:"send_to"` //client to send the message to
	WaitForResponse bool    `json:"wait_for_response"`
}

type CommandData struct {
	Command string `json:"command"`
	Output  string `json:"output"`
}
type StatsData struct {
	Name string `json:"name"`
}
type ErrorData struct {
	Error string `json:"error"`
}
