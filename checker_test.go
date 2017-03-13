package main

import (
	"fmt"
	"go/ast"
	"testing"
)

var src1 = `
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
	Name string ` + "`bson:\"name\"`" + `
	// Zip the company name
	Zip string ` + "`bson:\"zip_code\"`" + `
}

// User represents a user document in mongodb
// mgo:model:xyz_users
type User struct {
	// Name the user name
	Name string ` + "`bson:\"name\"`" + `
	// Email the uer's email
	Email string ` + "`bson:\"email\"`" + `
}

// connect is here
func connect() *mgo.Session {
	xyzWebSession, _ := mgo.DialWithTimeout("127.0.0.1:2700/dbname", 2*time.Second)
	return xyzWebSession
}
func findByName(name string) {
	var ret []Company
	testCollection := connect().DB("dbname").C("xyz_company")
	testCollection.Find(bson.M{"name": 1}).All(&ret)
}
`

var src2 = `
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
	Name string ` + "`bson:\"name\"`" + `
	// Zip the company name
	Zip string ` + "`bson:\"zip_code\"`" + `
}

// User represents a user document in mongodb
// mgo:model:xyz_users
type User struct {
	// Name the user name
	Name string ` + "`bson:\"name\"`" + `
	// Email the uer's email
	Email string ` + "`bson:\"email\"`" + `
}

// connect is here
func connect() *mgo.Session {
	xyzWebSession, _ := mgo.DialWithTimeout("127.0.0.1:2700/dbname", 2*time.Second)
	return xyzWebSession
}
func findByName(name string) {
	var ret []Company
	testCollection := connect().DB("dbname").C("xyz_company")
	testCollection.Find(bson.M{"names": "1"}).All(&ret)
}
`

func TestExpectStrGotInt(t *testing.T) {
	errorFound = nil
	collFieldTypes[`"xyz_company"."name"`] = "string"
	collFieldTypes[`"xyz_company"."zip_code"`] = "string"
	files := initCheckerSingleFile("sample1.go", "seeddata", src1)
	for _, f := range files {
		ast.Walk(&printASTVisitor{&info}, f)
		if errorFound.Actual != "int" {
			t.Errorf("actual: %s, expected: %s", errorFound.Actual, errorFound.Expected)
		}
	}

}

func TestWrongFieldName(t *testing.T) {
	errorFound = nil
	collFieldTypes[`"xyz_company"."name"`] = "string"
	collFieldTypes[`"xyz_company"."zip_code"`] = "string"
	files := initCheckerSingleFile("sample2.go", "seeddata", src2)
	for _, f := range files {
		ast.Walk(&printASTVisitor{&info}, f)
		fmt.Printf("%+v\n\n", errorFound)
		if errorFound.Expected != "" {
			t.Errorf("actual: %s, expected: %s", errorFound.Actual, errorFound.Expected)
		}
	}
}
