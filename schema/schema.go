package schema

import "time"

//User Schema
type User struct {
	ID                string    `json:"id" bson:"id" binding:"required"`
	Name              string    `json:"name" bson:"name" binding:"required"`
	DateOfBirth       string    `json:"dateofbirth" bson:"dateofbirth" binding:"required"`
	PhoneNumber       string    `json:"phonenumber" bson:"phonenumber" binding:"required"`
	EmailAddress      string    `json:"emailaddress" bson:"emailaddress" binding:"required"`
	CreationTimestamp time.Time `json:"creationtimestamp" bson:"creationtimestamp" binding:"required"`
}

//Contact Schema
type Contact struct {
	UserIDOne     string    `json:"useridone" bson:"useridone" binding:"required"`
	UserIDTwo     string    `json:"useridtwo" bson:"useridtwo" binding:"required"`
	TimeOfContact time.Time `json:"timeofcontact" bson:"timeofcontact" binding:"required"`
}
