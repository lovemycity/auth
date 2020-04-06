package user

import (
	"encoding/gob"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID primitive.ObjectID `json:"id" bson:"_id"`

	Email    string `json:"email" bson:"email"`
	Password string `json:"-" bson:"password"`

	FirstName  string `json:"first_name" bson:"first_name"`
	LastName   string `json:"last_name" bson:"last_name"`
	MiddleName string `json:"middle_name" bson:"middle_name"`

	DateOfBirth primitive.DateTime `json:"date_of_birth" bson:"date_of_birth"`

	Picture string `json:"picture" bson:"picture"`

	CreatedAt primitive.DateTime `json:"created_at" bson:"created_at"`
	CreatedIP string             `json:"created_ip" bson:"created_ip"`
}

func init() {
	gob.Register(new(User))
}
