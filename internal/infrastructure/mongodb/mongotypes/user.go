package mongotypes

type User struct {
	Id               string `bson:"_id"`
	TotalBalance     int64  `bson:"total_balance"`
	AvailableBalance int64  `bson:"available_balance"`
	Reserved         int64  `bson:"reserved"`
}
