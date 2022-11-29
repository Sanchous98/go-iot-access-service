package messages

import "time"

const UpdateNetworkType EventType = "updateNetworkState"

const (
	OpenState  NetworkState = "open"
	CloseState NetworkState = "close"
)

type NetworkState string

type UpdateNetworkState struct {
	Action   NetworkState  `json:"action"`
	Duration time.Duration `json:"duration"`
}

func NewUpdateNetworkState(transactionId int, state NetworkState) (event EventRequest[UpdateNetworkState]) {
	event.TransactionId = transactionId
	event.Event.EventType = UpdateNetworkType
	event.Event.Payload.Action = state

	if state == OpenState {
		event.Event.Payload.Duration = 10000
	}

	return
}
