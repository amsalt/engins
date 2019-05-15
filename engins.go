package engins

import "github.com/amsalt/nginet/message"

var Register message.Register
var Dispatcher message.ProcessorMgr

func init() {
	Register = message.NewRegister()
	Dispatcher = message.NewProcessorMgr(Register)
}

// GetMetaByID is a helper method by using default register.
func GetMetaByID(id interface{}) message.Meta {
	return Register.GetMetaByID(id)
}

// GetMetaByMsg is a helper method by using default register.
func GetMetaByMsg(msg interface{}) message.Meta {
	return Register.GetMetaByMsg(msg)
}

// RegisterMsg is a helper method by using default register.
func RegisterMsg(msg interface{}) (meta message.Meta) {
	return Register.RegisterMsg(msg)
}

// RegisterMsgByID is a helper method by using default register.
func RegisterMsgByID(assignID interface{}, msg interface{}) message.Meta {
	return Register.RegisterMsgByID(assignID, msg)
}

// RegisterProcessor is a helper method by using default Dispatcher.
func RegisterProcessor(msg interface{}, hf message.ProcessorFunc) error {
	return Dispatcher.RegisterProcessor(msg, hf)
}

// RegisterProcessorByID is a helper method by using default Dispatcher.
func RegisterProcessorByID(msgID interface{}, hf message.ProcessorFunc) error {
	return Dispatcher.RegisterProcessorByID(msgID, hf)
}

// GetProcessorByID is a helper method by using default Dispatcher.
func GetProcessorByID(msgID interface{}) *message.Processor {
	return Dispatcher.GetProcessorByID(msgID)
}
