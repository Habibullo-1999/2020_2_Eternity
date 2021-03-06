package api

import (
	"time"
)

type GetProfile struct {
	Id        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	BirthDate time.Time `json:"date"`
}

type SignUp struct {
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	Description string    `json:"description"`
	BirthDate   time.Time `json:"date"`
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateUser struct {
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	Description string    `json:"description"`
	BirthDate   time.Time `json:"date"`
}

type UpdatePassword struct {
	OldPassword string `json:"oldpassword"`
	NewPassword string `json:"newpassword"`
}

//type CreateComment struct {
//	IsRoot   bool   `json:"is_root"` // NOTE (Pavel S) if true, ParentId is not checked
//	ParentId int    `json:"parent_id"`
//	Content  string `json:"content"`
//	PinId    int    `json:"pin_id"`
//}
//
//type GetComment struct {
//	Id      int     `json:"id"`
//	Path    []int32 `json:"path"`
//	Content string  `json:"content"`
//	PinId   int     `json:"pin_id"`
//	UserId  int     `json:"user_id"`
//}

type UserAct struct {
	Username string `json:"username"`
}

type UserPage struct {
	Username    string `json:"username"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	//PinsList  []domain.PinResp `json:"pins_list"`
}

type CreateBoard struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type AttachDetachPin struct {
	BoardID int `json:"board_id"`
	PinID   int `json:"pin_id"`
}

type GetBoardPins struct {
	BoardID int `json:"board_id"`
	//LastPinId int `json:"pin_id"`
}

type IsFollowing struct {
	Following bool `json:"following"`
}
