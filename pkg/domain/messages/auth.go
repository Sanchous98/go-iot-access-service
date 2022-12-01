package messages

const AuthEvent EventType = "authEvent"

const (
	NoneType   AuthType = "none"
	NfcType    AuthType = "NFC"
	QrType     AuthType = "QR"
	MobileType AuthType = "Mobile"
	NumPadType AuthType = "numPad"
)

const (
	NoneStatus            AuthStatus = "none"
	SuccessOfflineStatus  AuthStatus = "succesOffline"
	FailedOfflineStatus   AuthStatus = "failedOffline"
	FailedPrivacyStatus   AuthStatus = "failedPrivacy"
	VerifyOnlineStatus    AuthStatus = "verifyOnline"
	FailedOnlineStatus    AuthStatus = "failedOnline"
	SuccessOnlineStatus   AuthStatus = "successOnline"
	ErrorTimeNotSetStatus AuthStatus = "errorTimeNotSet"
	NotFoundOfflineStatus AuthStatus = "NotFoundOffline"
	ErrorEncryptionStatus AuthStatus = "errorEncryption"
)

type AuthStatus string

type AuthType string

type Auth struct {
	HashKey    string     `json:"hashKey"`
	Timestamp  int        `json:"timestamp"`
	AuthType   AuthType   `json:"authType"`
	AuthStatus AuthStatus `json:"authStatus"`
	ChannelIds []int      `json:"channelIds,omitempty"`
}

func NewAuthEvent(transactionId int, auth Auth) (event EventRequest[Auth]) {
	event.Event.TransactionId = transactionId
	event.Event.EventType = AuthEvent
	event.Event.Payload = auth
	return
}
