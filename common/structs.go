package common

type Message struct {
	Data interface{} `json:"data"`
}
type CommandData struct {
	Command string `json:"command"`
	Output  string `json:"output"`
}
type StatsData struct {
	Name string `json:"name"`
}
