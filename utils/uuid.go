package utils

import "github.com/google/uuid"

type UUID struct {
	uuid.UUID
}

func (u *UUID) UnmarshalJSON(id []byte) (err error) {
	if BytesToStr(id) == "null" {
		return
	}

	u.UUID, err = uuid.Parse(BytesToStr(id))
	return
}
