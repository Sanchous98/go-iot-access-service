package messages

const (
	LockActionOpen      EventType = "lockActionOpen"
	LockActionClose     EventType = "lockActionClose"
	LockActionAuto      EventType = "lockActionAuto"
	LockActionResponse  EventType = "lockActionResponse"
	LockOfflineResponse EventType = "lockOfflineResponse"
)

const (
	NoneLockStatus                   LockStatus = "none"
	ExtRelayStateLockStatus          LockStatus = "extRelayState"
	LockOpenedLockStatus             LockStatus = "lockOpened"
	LockClosedLockStatus             LockStatus = "lockClosed"
	DriverOnLockStatus               LockStatus = "driverOn"
	ErrorLockAlreadyOpenLockStatus   LockStatus = "errorLockAlreadyOpen"
	ErrorLockAlreadyClosedLockStatus LockStatus = "errorLockAlreadyClosed"
	ErrorDriverEnabledLockStatus     LockStatus = "errorDriverEnabled"
	DeviceTypeUnknownLockStatus      LockStatus = "deviceTypeUnknown"
	OfflineTimeoutLockStatus         LockStatus = "openTimeoutError"
)

type LockStatus string

type LockAuto struct {
	RecloseDelay byte `json:"recloseDelay"`
	LockResponse `json:"-"`
}

type LockResponse struct {
	LockActionStatus LockStatus `json:"lockActionStatus"`
	ChannelIds       []int      `json:"channelIds,omitempty"`
}

func NewLockOpenEvent(transactionId int, channelIds []int) (event EventRequest[LockAuto]) {
	event.TransactionId = transactionId
	event.Event.EventType = LockActionOpen
	event.Event.Payload.ChannelIds = channelIds
	return
}

func NewLockCloseEvent(transactionId int) (event EventRequest[EmptyPayload]) {
	event.TransactionId = transactionId
	event.Event.EventType = LockActionClose
	return
}

func NewLockAutoEvent(transactionId int, delay byte, channelIds []int) (event EventRequest[LockAuto]) {
	event.TransactionId = transactionId
	event.Event.EventType = LockActionAuto
	event.Event.Payload.RecloseDelay = delay
	event.Event.Payload.ChannelIds = channelIds
	return
}

func NewLockOfflineResponse() (event EventResponse[LockResponse]) {
	event.Event.EventType = LockOfflineResponse
	event.Event.Payload.LockActionStatus = OfflineTimeoutLockStatus
	return
}
