package models

type UID string

func (u UID) String() string {
	return string(u)
}
