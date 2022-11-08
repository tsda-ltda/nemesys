package tools

const (
	MsgIdentIsNumber            = "Identification must not be number as text."
	MsgRequestTimeout           = "Request timeout."
	MsgParamsNotSameType        = "Params must have same type. Use only numbers or only text."
	MsgInvalidParams            = "Invalid route params."
	MsgInvalidBody              = "Invalid body."
	MsgInvalidJSONFields        = "Invalid JSON fields."
	MsgIdentExists              = "Identification already exists."
	MsgTargetPortExists         = "Target and port combination already exists."
	MsgMaxDataPolicy            = "Max number of data policies reached."
	MsgWrongUsernameOrPW        = "Wrong username or password."
	MsgUsernameExists           = "Username already exists."
	MsgEmailExists              = "Email already exists."
	MsgContainerNotFound        = "Container does not exists."
	MsgMetricNotFound           = "Metric does not exists."
	MsgDataPolicyNotFound       = "Data policy does not exists."
	MsgTeamNotFound             = "Team does not exists."
	MsgContextNotFound          = "Context does not exists."
	MsgContextualMetricNotFound = "Contextual metric does not exists."
	MsgUserNotFound             = "User does not exists."
	MsgSessionAlreadyRemoved    = "Session already removed."
)

type Message struct {
	Message string `json:"message"`
}

func JSONMSG(msg string) Message {
	return Message{
		Message: msg,
	}
}
