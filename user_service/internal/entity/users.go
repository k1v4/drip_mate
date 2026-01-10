package entity

import "encoding/json"

type User struct {
	ID            int    `json:"id"`
	Email         string `json:"email"`
	Password      []byte `json:"-"`
	Name          string `json:"name"`
	Surname       string `json:"surname"`
	Username      string `json:"username"`
	AccessLevelId int    `json:"-"`
}

func (o *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(o)
}

func (o *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, o)
}
