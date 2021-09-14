package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role int

const (
	GENERAL_USER Role = 0
	ADMIN        Role = 1
)

// User -
type User struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Firstname  string             `json:"firstname" bson:"firstname" validate:"empty=false"`
	Lastname   string             `json:"lastname" bson:"lastname" validate:"empty=false"`
	Email      string             `json:"email" bson:"email" validate:"empty=false & format=email"`
	Password   string             `json:"password" bson:"password" validate:"empty=false"`
	UserRole   Role               `json:"role" bson:"role" validate:"empty=false"`
	IsVerified bool               `json:"isverified" bson:"isverified" validate:"empty=false"`
	CreatedAt  int64              `json:"created_at" bson:"created_at" validate:"empty=false"`
	UpdatedAt  int64              `json:"updated_at" bson:"updated_at"`
}
