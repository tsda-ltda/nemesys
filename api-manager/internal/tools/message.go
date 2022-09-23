package tools

type Message struct {
	Message string `json:"message"`
}

func NewMsg(msg string) Message {
	return Message{
		Message: msg,
	}
}
