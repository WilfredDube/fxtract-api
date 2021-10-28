package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// Project -
type Project struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title" validate:"empty=false"`
	Description string             `json:"description" bson:"description" validate:"empty=false"`
	OwnerID     primitive.ObjectID `json:"-" bson:"owner_id" validate:"empty=false"`
	CreatedAt   int64              `json:"created_at" bson:"created_at" validate:"empty=false"`
}
