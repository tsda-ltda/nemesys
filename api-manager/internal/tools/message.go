package tools

const (
	MsgIdentExists        = "Identification already exists."
	MsgTargetPortExists   = "Target and port combination already exists."
	MsgMaxDataPolicy      = "Max number of data policies reached."
	MsgWrongUsernameOrPW  = "Wrong username or password."
	MsgUsernameExists     = "Username already exists."
	MsgEmailExists        = "Email already exists."
	MsgContainerNotFound  = "Container does not exists."
	MsgDataPolicyNotFound = "Data policy does not exists."
)

type Message struct {
	Message string `json:"message"`
}

func JSONMSG(msg string) Message {
	return Message{
		Message: msg,
	}
}
