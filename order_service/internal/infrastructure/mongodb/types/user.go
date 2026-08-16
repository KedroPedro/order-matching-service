package types

import "github.com/KedroPedro/order-matching-engine/order_service/internal/domain/entity"

type User struct {
	Id               string `bson:"_id"`
	Login            string `bson:"login"`
	TotalBalance     int64  `bson:"total_balance"`
	AvailableBalance int64  `bson:"available_balance"`
	Reserved         int64  `bson:"reserved"`
}

func FromDomainEntity(user *entity.User) *User {
	return &User{
		Id:               user.Id,
		TotalBalance:     user.TotalBalance,
		AvailableBalance: user.AvailableBalance,
		Reserved:         user.Reserved,
		Login:            user.Login,
	}
}

func (this *User) ToDomainEntity() *entity.User {
	return &entity.User{
		Id:               this.Id,
		TotalBalance:     this.TotalBalance,
		AvailableBalance: this.AvailableBalance,
		Reserved:         this.Reserved,
		Login:            this.Login,
	}
}
