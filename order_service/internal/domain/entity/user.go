package entity

type User struct {
	Id               string
	Login            string
	TotalBalance     int64
	AvailableBalance int64
	Reserved         int64
}
