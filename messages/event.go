package messages

type EventType string

type EmptyPayload struct{}

type EventRequestPayloads interface {
	DeviceConfig | EmptyPayload | LockAuto | LocalStorageEvent | Auth
}

type EventRequest[T EventRequestPayloads] struct {
	Event struct {
		EventType EventType `json:"eventType"`
		Payload   T         `json:"payload"`
	} `json:"event"`
	TransactionId int `json:"transactionId"`
}

type EventResponse[T LockResponse | DeviceStatusResponsePayload] struct {
	Event struct {
		ShortAddr string    `json:"short_addr"`
		ExtAddr   string    `json:"ext_addr"`
		Rssi      int       `json:"rssi"`
		EventType EventType `json:"eventType"`
		Payload   T         `json:"payload"`
	} `json:"event"`
	TransactionId int `json:"transactionId"`
}

type Response[T Auth] struct {
	ShortAddr     string    `json:"short_addr"`
	ExtAddr       string    `json:"ext_addr"`
	Rssi          int       `json:"rssi"`
	EventType     EventType `json:"eventType"`
	Payload       T         `json:"payload"`
	TransactionId int       `json:"transactionId"`
}
