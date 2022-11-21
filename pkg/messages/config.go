package messages

const (
	DeviceConfigRead   EventType = "deviceConfigRead"
	DeviceConfigUpdate EventType = "deviceConfigUpdate"
)

type DeviceConfig struct {
	TxPower                 uint   `json:"txPower,omitempty"`
	DeviceType              string `json:"deviceType,omitempty"`
	DeviceRole              string `json:"deviceRole,omitempty"`
	FrontBreakout           string `json:"frontBreakout,omitempty"`
	BackBreakout            string `json:"backBreakout,omitempty"`
	RecloseDelay            uint   `json:"recloseDelay,omitempty"`
	StatusMsgFlags          uint   `json:"statusMsgFlags,omitempty"`
	StatusUpdateInterval    uint16 `json:"statusUpdateInterval,omitempty"`
	NfcEncryptionKey        string `json:"nfcEncryptionKey,omitempty"`
	InstalledRelayModuleIds []uint `json:"installedRelayModuleIds,omitempty"`
	ExternalRelayMode       string `json:"externalRelayMode,omitempty"`
	SlaveFwAddress          uint   `json:"slaveFwAddress,omitempty"`
	BuzzerVolume            string `json:"buzzerVolume,omitempty"`
	EmvCoPrivateKey         string `json:"emvCoPrivateKey,omitempty"`
	EmvCoKeyVersion         string `json:"emvCoKeyVersion,omitempty"`
	EmvCoCollectorId        string `json:"emvCoCollectorId,omitempty"`
	GoogleSmartTapEnabled   bool   `json:"googleSmartTapEnabled,omitempty"`
}
