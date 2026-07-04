package models

import "time"

type FCMTokenModel struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	Token      string    `json:"token"`
	Plataforma string    `json:"plataforma"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
