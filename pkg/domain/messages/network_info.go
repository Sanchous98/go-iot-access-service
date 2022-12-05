package messages

const GetNetworkInfoType EventType = "updateNetworkState"

func NewNetworkInfoRequest(transactionId int) (event EventRequest[EmptyPayload]) {
	event.Event.TransactionId = transactionId
	event.Event.EventType = GetNetworkInfoType
	return
}
