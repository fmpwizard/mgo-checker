package seeddata

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

// Company represents a Company document in mongodb
// mgo:model:xyz_company
type Company struct {
	// Name the company name
	Name string `bson:"name"`
	// Zip the company name
	Zip string `bson:"zip_code"`
}

// User represents a user document in mongodb
// mgo:model:xyz_users
type User struct {
	// Name the user name
	Name string `bson:"name"`
	// Email the uer's email
	Email string `bson:"email"`
}

// connect is here
func connect() *mgo.Session {
	xyzWebSession, _ := mgo.DialWithTimeout("127.0.0.1:2700/dbname", 2*time.Second)
	return xyzWebSession
}
func findByName(name string) {
	var ret []Company
	collName := "xyz_company"
	testCollection := connect().DB("dbname").C(collName)
	testCollection.Find(bson.M{"name": 1}).All(&ret)
}
