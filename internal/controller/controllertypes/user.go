package controllertypes

import (
	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/google/uuid"
)

type User struct {
	Id               string `json:"id,omitempty"`
	Login            string `json:"login,omitempty"`
	TotalBalance     int64  `json:"total_balance,omitempty"`
	AvailableBalance int64  `json:"available_balance,omitempty"`
	Reserved         int64  `json:"reserved,omitempty"`
}

func FromDomainEntity(user *entity.User) *User {
	return &User{
		Id:               user.Id,
		Login:            user.Login,
		TotalBalance:     user.TotalBalance,
		AvailableBalance: user.AvailableBalance,
		Reserved:         user.Reserved,
	}
}

func (this *User) ToDomainEntity() *entity.User {
	if this.Id == "" {
		this.Id = uuid.NewString()
	}

	return &entity.User{
		Id:               this.Id,
		Login:            this.Login,
		TotalBalance:     this.TotalBalance,
		AvailableBalance: this.AvailableBalance,
		Reserved:         this.Reserved,
	}
}
