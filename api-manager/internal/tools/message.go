package tools

const (
	MsgContainerNotFound        = "Container does not exists."
	MsgMetricNotFound           = "Metric does not exists."
	MsgDataPolicyNotFound       = "Data policy does not exists."
	MsgTeamNotFound             = "Team does not exists."
	MsgContextNotFound          = "Context does not exists."
	MsgContextualMetricNotFound = "Contextual metric does not exists."
	MsgUserNotFound             = "User does not exists."
	MsgMemberNotFound           = "Member does not exists."
	MsgCustomQueryNotFound      = "Custom query does not exists."

	MsgParamsNotSameType     = "Params must have same type. Use only numbers or only text."
	MsgIdentIsNumber         = "Identification must not be number as text."
	MsgRequestTimeout        = "Request timeout."
	MsgMaxDataPolicy         = "Max number of data policies reached."
	MsgWrongUsernameOrPW     = "Wrong username or password."
	MsgSessionAlreadyRemoved = "Session already removed."
	MsgMetricDisabled        = "Metric is not enabled."
	MsgContainerDisabled     = "Container is not enabled."

	MsgInvalidParams     = "Invalid route params."
	MsgInvalidBody       = "Invalid body."
	MsgInvalidJSONFields = "Invalid JSON fields."

	MsgIdentExists        = "Identification already exists."
	MsgTargetPortExists   = "Target and port combination already exists."
	MsgEmailExists        = "Email already exists."
	MsgUsernameExists     = "Username already exists."
	MsgRelationExists     = "User is already a member."
	MsgSerialNumberExists = "Flex serial-number already exists. "
)

type Message struct {
	Message string `json:"message"`
}

func JSONMSG(msg string) Message {
	return Message{
		Message: msg,
	}
}
