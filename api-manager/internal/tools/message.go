package tools

type Message struct {
	Message string `json:"message"`
}

func JSONMSG(msg string) Message {
	return Message{
		Message: msg,
	}
}
