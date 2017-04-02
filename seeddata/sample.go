package seeddata

import (
	"fmt"
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

func searchStep1() {
	companyColl := connect().DB("dbname").C("xyz_company")
	fmt.Println("some random line her")
	fmt.Println("another here, just tomess with you")
	findByZip(companyColl, 1)
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

func findByZip(collection *mgo.Collection, n int64) {
	var ret []Company
	err := collection.Find(bson.M{"name": n}).All(&ret)
	_ = err
}
