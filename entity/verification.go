package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VerificationDataType int

const (
	MailConfirmation VerificationDataType = iota + 1
	PassReset
)

type Verification struct {
	ID        primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Email     string               `json:"email" validate:"required" bson:"email"`
	Code      string               `json:"code" validate:"required" bson:"code"`
	ExpiresAt int64                `json:"expires_at" validate:"required" bson:"expires_at"`
	Type      VerificationDataType `json:"type" validate:"required" bson:"type"`
}
