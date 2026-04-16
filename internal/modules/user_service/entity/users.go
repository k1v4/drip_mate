package entity

import "encoding/json"

type User struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	Password        string `json:"-"`
	Name            string `json:"name"`
	Surname         string `json:"surname"`
	Username        string `json:"username"`
	City            string `json:"city"`
	AccessLevelName string `json:"access_level"`
	AccessLevelId   int    `json:"-"`
}

func (o *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(o)
}

func (o *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, o)
}
