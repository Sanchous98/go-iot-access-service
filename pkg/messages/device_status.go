package messages

const (
	DeviceStatusRequest  EventType = "deviceStatusReq"
	DeviceStatusResponse EventType = "deviceStatusRsp"
)

const (
	ErrorDetectedReason Reason = "errorDetected"
)

type Reason string

type DeviceStatusResponsePayload struct {
	Reason     Reason `json:"reason"`
	Timezone   int    `json:"timezone"`
	LockSensor struct {
		Raw     int `json:"raw"`
		Privacy int `json:"privacy"`
		Handle  int `json:"handle"`
		Key     int `json:"key"`
	} `json:"lockSensor"`
}
