package messages

const LocateRequest EventType = "locateReq"

func NewLocateRequest(transactionId int) (event EventRequest[EmptyPayload]) {
	event.Event.TransactionId = transactionId
	event.Event.EventType = LocateRequest
	return
}
