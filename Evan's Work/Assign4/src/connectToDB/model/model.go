// authors Jordan Lys & Evan Closson

package model

import (
 	//"gopkg.in/mgo.v2"
  	"gopkg.in/mgo.v2/bson"
	"time"
	)

type User struct {
	ID        	bson.ObjectId 	`bson:"_id"`
	Name      	string			`bson:"name"`
	Phone     	string			`bson:"phone"`
	Email		string			`bson:"email"`
	IsRealUser	bool			`bson:"isUser"`
	Groups		[]Group 		`bson:"groups"`
	Contacts	[]Contact		`bson:"contacts"`
	Timestamp 	time.Time 		`bson:"time.Time"`
}

type Contact struct {
	ID 			bson.ObjectId 	`bson:"_id"`
	Name		string 			`bson:"name"`
	Phone		string  		`bson:"phone"`
	Email		string 			`bson:"email"`

}

type Group struct {
	ID 			bson.ObjectId 	`bson:"_id"`
	GroupName	string 			`bson:"groupName"`
	Users 		[]User			`bson:"users"`
}

type Comment struct {
	ID 			bson.ObjectId 	`bson:"_id"`
	UserName	string 			`bson:"userName"`
	Subject		string 			`bson:"subject"`
	Content		string 			`bson:"content"`
	Timestamp	time.Time  		`bson:"time.Time"`
}


