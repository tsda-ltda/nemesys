package tools

import "github.com/fernandotsda/nemesys/api-manager/internal/api"

const (
	MsgContainerNotFound                   = "Container does not exists."
	MsgMetricNotFound                      = "Metric does not exists."
	MsgDataPolicyNotFound                  = "Data policy does not exists."
	MsgTeamNotFound                        = "Team does not exists."
	MsgContextNotFound                     = "Context does not exists."
	MsgContextualMetricNotFound            = "Contextual metric does not exists."
	MsgUserNotFound                        = "User does not exists."
	MsgMemberNotFound                      = "Member does not exists."
	MsgCustomQueryNotFound                 = "Custom query does not exists."
	MsgRefkeyNotFound                      = "Metric reference not exists."
	MsgAPIKeyNotFound                      = "API Key does not exists."
	MsgAlarmExpressionNotFound             = "Alarm expression does not exists."
	MsgAlarmProfileNotFound                = "Alarm profile does not exists."
	MsgUserWhitelistNotFound               = "User does not exists in whitelist."
	MsgAlarmCategoryNotFound               = "Alarm category does not exists."
	MsgAlarmProfileAndCategoryRelNotFound  = "Alarm profile and alarm category relation does not exists."
	MsgAlarmExpressionAndMetricRelNotFound = "Alarm expression and metric relation does not exists."
	MsgAlarmProfileEmailNotFound           = "Alarm profile email does not exists."
	MsgTrapRelationNotFound                = "Trap category relation does not exists."
	MsgAlarmStateNotFound                  = "Alarm state does not exists."
	MsgTrapListenerNotFound                = "Trap listener does not exists."

	MsgParamsNotSameType     = "Params must have same type. Use only numbers or only text."
	MsgIdentIsNumber         = "Identification must not be number as text."
	MsgRequestTimeout        = "Request timeout."
	MsgMaxDataPolicy         = "Max number of data policies reached."
	MsgWrongUsernameOrPW     = "Wrong username or password."
	MsgSessionAlreadyRemoved = "Session already removed."
	MsgMetricDisabled        = "Metric is not enabled."
	MsgContainerDisabled     = "Container is not enabled."
	MsgMetricIsNotAlarmed    = "Metric alarm state is not alarmed."
	MsgMetricIsNotRecognized = "Metric alarm state is not recognized."

	MsgInvalidParams     = "Invalid route params."
	MsgInvalidBody       = "Invalid body."
	MsgInvalidMetricType = "Invalid metric type."
	MsgInvalidAggrFn     = "Invalid data-policy aggregation function."
	MsgInvalidJSONFields = "Invalid JSON fields."
	MsgInvalidMetricData = "Invalid metric data, could not parse input data to metric type. Check if metric type is correct."
	MsgInvalidRole       = "Invalid user role."

	MsgIdentExists                       = "Identification already exists."
	MsgTargetPortExists                  = "Target and port combination already exists."
	MsgTargetExists                      = "Target already exists."
	MsgEmailExists                       = "Email already exists."
	MsgUsernameExists                    = "Username already exists."
	MsgRelationExists                    = "User is already a member."
	MsgSerialNumberExists                = "Flex serial-number already exists."
	MsgRefkeyExists                      = "Metric reference key already exists."
	MsgAlarmExpressionExists             = "Alarm expression already exists."
	MsgAlarmCategoryLevelExists          = "Alarm category level already exists."
	MsgAlarmProfileAndCategoryRelExists  = "Alarm profile and alarm category relation already exists."
	MsgAlarmExpressionAndMetricRelExists = "Alarm expression and metric relation already exists."
	MsgTrapRelationExists                = "Trap category already have a relation."
	MsgTrapListerHostPortExists          = "Trap listener host port already exists."
)

// MsgRes returns an APIResponse with empty data but with the message.
func MsgRes(msg string) (response api.APIResponse) {
	response.Message = msg
	return response
}

// EmptyRes returns an empty APIResponse.
func EmptyRes() (response api.APIResponse) {
	return response
}

type idMSG struct {
	Id int64 `json:"id"`
}

func DataRes(data any) (response api.APIResponse) {
	response.Data = data
	return response
}

// IdRes return and APIResponse with empty message but and id as data.
func IdRes(id int64) (response api.APIResponse) {
	response.Data = idMSG{Id: id}
	return response
}
