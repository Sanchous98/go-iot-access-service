package messages

const RemoveDeviceType EventType = "removeDeviceReq"

type RemoveDevice struct {
	DeviceMac string `json:"extAddress"`
}

func NewRemoveDeviceRequest(transactionId int, deviceMacId string) (event EventRequest[RemoveDevice]) {
	event.Event.TransactionId = transactionId
	event.Event.EventType = RemoveDeviceType
	event.Event.Payload.DeviceMac = deviceMacId
	return
}
