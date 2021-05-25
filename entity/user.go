package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// User -
type User struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname" bson:"firstname" validate:"empty=false"`
	Lastname  string             `json:"lastname" bson:"lastname" validate:"empty=false"`
	Email     string             `json:"email" bson:"email" validate:"empty=false & format=email"`
	Password  string             `json:"password" bson:"password" validate:"empty=false"`
	Token     string             `json:"token" bson:"token" validate:"empty=false"`
	CreatedAt int64              `json:"created_at" bson:"created_at" validate:"empty=false"`
}
