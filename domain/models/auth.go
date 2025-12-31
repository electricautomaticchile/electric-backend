package models

type LoginResponseModel struct {
	Token        string        `json:"token"`
	RefreshToken string        `json:"refreshToken"`
	User         *ClienteModel `json:"user"`
}
