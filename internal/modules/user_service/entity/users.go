package entity

import (
	"encoding/json"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	City        string    `json:"city"`
	AccessID    int       `json:"access_id"`
	AccessLevel string    `json:"access_level"`

	Music  []string `json:"music"`
	Styles []string `json:"styles"`
	Colors []string `json:"colors"`

	Outfits []Outfit `json:"outfits"`
}

type UpdatePersonal struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Username string `json:"username"`
	Gender   string `json:"gender"`
}

type UpdatePass struct {
	CurrPassword string `json:"curr_password" validate:"required"`
	NewPassword  string `json:"new_password" validate:"required,min=10"`
}

func (o *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(o)
}

func (o *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, o)
}
