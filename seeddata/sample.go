package seeddata

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

// Company represents a Company document in mongodb
// mgo:model
type Company struct {
	// Name the company name
	Name string
	// Zip the company name
	Zip string
}

// connect is here
func connect() *mgo.Session {
	acmWebSession, err := mgo.DialWithTimeout("192.168.1.11:2700/acm-web", 2*time.Second)
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %q", err)
	}
	return acmWebSession
}
func findByName(name string) {
	var ret []Company
	testCollection := connect().DB("acm-web").C("test")
	testCollection.Find(bson.M{"name": 1}).All(&ret)
}

func do() {
	findByName("diego")
}
