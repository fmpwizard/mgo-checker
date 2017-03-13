package seeddata

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Company represents a Company document in mongodb
// mgo:model:xyz_company
type Company struct {
	// Name the company name
	Name string `bson:"name"`
	// Zip the company zip code
	Zip string `bson:"zip_code"`
}

// connect is here
func connect() *mgo.Session {
	xyzWebSession, _ := mgo.DialWithTimeout("127.0.0.1:2700/dbname", 0)
	return xyzWebSession
}
func findByName(name string) {
	var ret []Company
	testCollection := connect().DB("dbname").C("xyz_company")
	testCollection.Find(bson.M{"name": 1}).All(&ret)
}
