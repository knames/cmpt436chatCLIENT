package main

import (
 	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
	"fmt"
)


type User struct {
	ID        	bson.ObjectId 	`json:"id" bson:"_id,omitempty"`
	Name      	string			`json:"name"`
	Phone     	string			`json:"phone"`
	Email		string			`json:"email"`
	IsRealUser	bool			`json:"isRealUser"`
	Groups		[]*Group 		`json:"groups" bson:"groups"`
	Contacts	[]*Contact		`json:"contacts" bson:"contacts"`
	Timestamp 	time.Time 		`json:"time"`
}

type Contact struct {
	ID 			bson.ObjectId 	`json:"id" bson:"_id,omitempty"`
	Name		string 			`json:"name"`
	Phone		string  		`json:"phone"`
	Email		string 			`json:"email"`
	IsRealUser	bool			`json:"isRealUser`
	Timestamp 	time.Time 		`json:"time"`
}

type Group struct {
	ID 			bson.ObjectId 	`json:"id" bson:"_id, omitempty"`
	GroupName	string 			`json:"groupName"`
	Users 		[]*User			`json:"users"`
	Expected	[]*int 			`json:"expected"`
	Actual		[]*int 			`json:"actual"`
}

type Comment struct {
	ID 			bson.ObjectId 	`json:"id" bson:"_id, omitempty"`
	UserName	string 			`json:"userName"`
	Subject		string 			`json:"subject"`
	Content		string 			`json:"content"`
	Timestamp	time.Time  		`json:"time"`
}

var (
	IsDrop = true
	session *mgo.Session
	collection *mgo.Database
)
func thisPanic(err error) {
  if err != nil {
    panic(err)
  }

}

func connectToDB() {

	 
	 var err error
	 session, err = mgo.Dial("127.0.0.1")
	 thisPanic(err)
	 collection = session.DB("")

}
func main() {
	var err error
	connectToDB()
	thisPanic(err)

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	// Drop Database
	if IsDrop {
		err = session.DB("test").DropDatabase()
		thisPanic(err)

	}
	c := session.DB("test").C("User")

	
	index := mgo.Index {
		Key:        []string{"name", "phone"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)

	thisPanic(err)


	err = c.Insert(&User{ Name: "Ale", Phone: "+922", IsRealUser: true, Email:"abc@gmail.com", Timestamp: time.Now()})
	thisPanic(err)
	err = c.Insert(&User{ Name: "Jrock", Phone: "+911", IsRealUser: true, Email:"jcl@gmail.com", Timestamp: time.Now()})	
	thisPanic(err)


	c = session.DB("test").C("Contact")
	err = c.Insert(&Contact{ Name: "Ale", Phone: "+922", IsRealUser: true, Email:"abc@gmail.com", Timestamp: time.Now()})
	thisPanic(err)



	result := Contact{}
	err = c.Find(bson.M{"name": "Ale"}).One(&result)
	thisPanic(err)

	fmt.Println("\n")
	fmt.Println(result)
	fmt.Println("\n")

	findJ := User{}
	c = session.DB("test").C("User")
	err = c.Find(bson.M{"name": "Jrock"}).Select(bson.M{"_id":1}).One(&findJ)
	fmt.Println(findJ)
	thisPanic(err)

	fmt.Println("\nHexID of JRock\n")
	fmt.Println(findJ.ID.Hex())
	fmt.Println("\nResult Object\n")
	fmt.Println(result)

	 query := bson.M{"_id": bson.ObjectId(findJ.ID)}
	 fmt.Println("\nQuery\n")
	 fmt.Println(query)
	 change := bson.M{"$push": bson.M{"contacts": result}}
	 //change2 := bson.M{"$push": bson.M{"contacts": bson.M{"name": result.Name}}}

	 fmt.Println("\nUpdate Params\n")
	 fmt.Println(change)
	 err = c.Update(query, change)
	 thisPanic(err)

	 findJ = User{}
	 err = c.Find(bson.M{"name": "Jrock"}).One(&findJ)
	 thisPanic(err)

	 fmt.Println("\nContacts of JRock\n")
	 fmt.Println(findJ.Contacts[0])
	 
	c = session.DB("test").C("Contact")
	err = c.Insert(&Contact{ Name: "Eclo", Phone: "+306", IsRealUser: true, Email:"eclo@gmail.com", Timestamp: time.Now()})
	thisPanic(err)
	result = Contact{}
	err = c.Find(bson.M{"name": "Eclo"}).One(&result)
	
	c = session.DB("test").C("User")
	/*ADD ANOTHER CONTACT*/
	findJ = User{}
	c = session.DB("test").C("User")
	err = c.Find(bson.M{"name": "Jrock"}).Select(bson.M{"_id":1}).One(&findJ)
	fmt.Println(findJ)
	thisPanic(err)

	fmt.Println("\nHexID of JRock\n")
	fmt.Println(findJ.ID.Hex())
	fmt.Println("\nResult Object\n")
	fmt.Println(result)

	 query = bson.M{"_id": bson.ObjectId(findJ.ID)}
	 fmt.Println("\nQuery\n")
	 fmt.Println(query)
	 change = bson.M{"$push": bson.M{"contacts": result}}
	 //change2 := bson.M{"$push": bson.M{"contacts": bson.M{"name": result.Name}}}

	 fmt.Println("\nUpdate Params\n")
	 fmt.Println(change)
	 err = c.Update(query, change)
	 thisPanic(err)

	 findJ = User{}
	 err = c.Find(bson.M{"name": "Jrock"}).One(&findJ)
	 thisPanic(err)

	 fmt.Println("\nContacts of JRock\n")
	 fmt.Println(findJ.Contacts[1])
	 


}	
