package messages

const (
	LocalStorageAddKey    EventType = "localStorageAddKey"
	LocalStorageReadKey   EventType = "localStorageGetKey"
	LocalStorageDeleteKey EventType = "localStorageDeleteKey"
)

const (
	OkLocalStorageStatus LocalStorageStatusCodes = iota
	OkLocalStorageRead
	ErrorKeyNotFound
	ErrorKeyAlreadyExists
	ErrorFlashStorageIsFull
	ErrorCritical
)

type LocalStorageStatusCodes int

type LocalStorageUpdateKeys struct {
	HashKey string `json:"hashKey"`
	Flags   struct {
		MasterKey            bool `json:"masterKey"`
		PrivacyOverride      bool `json:"privacyOverride"`
		IsMultiChannel       bool `json:"isMultiChannel"`
		IsMeetingModeAllowed bool `json:"isMeetingModeAllowed"`
	} `json:"flags"`
	MasterKey struct {
		ChannelIds []int `json:"channelIds"`
	} `json:"masterKey"`
	TimeKeys []struct {
		StartTime  string `json:"startTime"`
		EndTime    string `json:"endTime"`
		ChannelIds []int  `json:"channelIds"`
	} `json:"timeKeys"`
	AclKeys []struct {
		DaysOfWeek []int  `json:"daysOfWeek"`
		StartTime  string `json:"startTime"`
		EndTime    string `json:"endTime"`
		ChannelIds []int  `json:"channelIds"`
	} `json:"aclKeys"`
}

type LocalStorageReadKeys struct {
	HashKey string `json:"hashKey"`
}

type LocalStorageDeleteKeys struct {
	HashKey string `json:"hashKey"`
}
