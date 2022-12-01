package messages

const FirmwareVersionRequest EventType = "fwVersionReq"
const UpdateFirmwareVersionRequest EventType = "fwUpdateReq"

func NewFirmwareVersionRequest(transactionId int) (event EventRequest[EmptyPayload]) {
	event.TransactionId = transactionId
	event.Event.EventType = FirmwareVersionRequest
	return
}
