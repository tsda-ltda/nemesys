package amqp

type MessageType uint16

const (
	// MTUntyped is sent when a message is generic. Can vary
	// according to the queue contexts.
	MTUntyped MessageType = iota
	// MTDataMismatch is sent when a previous message have
	// data that doesn't match with it's type.
	MTDataMismatch
	// MTInternalError is sent when something went wrong,
	// while processing the message.
	MTInternalError
	// MTInvalidBody is sent when the message body could be
	// unmarshed correctly.
	MTInvalidBody
)
