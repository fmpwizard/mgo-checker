package main

import (
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

var src3 = `
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
	Name string
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

var src4 = `
package seeddata


// Company represents a Company document in mongodb
// mgo:model:xyz_company
type Company struct {
	// Name the company name
	Name, Street string
	// Zip the company name
	Zip string ` + "`bson:\"zip_code\"`" + `
}
`

var src5 = `
package seeddata


// Company represents a Company document in mongodb
// mgo:model:xyz_company
type Company struct {
	// Name the company name
	Name, Street string
	// Zip the company name
	Zip string ` + "`json:\"zip_code\" bson:\"zip_code\"`" + `
}
`

var src6 = `
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
	Name string
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
	testCollection.Find(bson.M{"name": "1", "zip_code": 1}).All(&ret)
}
`

var src7 = `
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
	Name string
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
	testCollection.Find(bson.M{"name": "1", "zip": "23445"}).All(&ret)
}
`

func TestExpectStrGotInt(t *testing.T) {
	errorFound = nil
	collFieldTypes = make(map[string]string)
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
	collFieldTypes = make(map[string]string)
	files := initCheckerSingleFile("sample2.go", "seeddata", src2)
	for _, f := range files {
		ast.Walk(&printASTVisitor{&info}, f)
		if errorFound.Expected != "" {
			t.Errorf("actual: %s, expected: %s", errorFound.Actual, errorFound.Expected)
		}
	}
}

func TestGetMgoCollectionFromComment(t *testing.T) {
	found, name := getMgoCollectionFromComment(`Company represents a Company document in mongodb
mgo:model:zzz_company`)
	if !found {
		t.Error("didn't find direective")
	}
	if name != "zzz_company" {
		t.Errorf("expected `zzz_company` but got: `%s`", name)
	}
}

func TestFieldFromTag1(t *testing.T) {
	r := fieldFromTag("`bson:\"name\"`")
	if r != "name" {
		t.Errorf("expeted `name` but got: %s", r)
	}
}

func TestFieldFromTag2(t *testing.T) {
	r := fieldFromTag("name")
	if r != "name" {
		t.Errorf("expeted `name` but got: %s", r)
	}
}

func TestFieldFromTag3(t *testing.T) {
	r := fieldFromTag("name,omitempty")
	if r != "name" {
		t.Errorf("expeted `name` but got: %s", r)
	}
}

func TestFieldFromTag4(t *testing.T) {
	r := fieldFromTag("`bson:\",omitempty\" json:\"jsonkey\"`")
	if r != "" {
		t.Errorf("expeted `<empty string>` but got: %s", r)
	}
}

func TestFieldFromTag5(t *testing.T) {
	r := fieldFromTag(",minsize")
	if r != "" {
		t.Errorf("expeted `<empty string>` but got: %s", r)
	}
}

func TestFieldFromTag6(t *testing.T) {
	r := fieldFromTag("name,omitempty,minsize")
	if r != "name" {
		t.Errorf("expeted `name` but got: %s", r)
	}
}

func TestFieldFromTag7(t *testing.T) {
	r := fieldFromTag("`json:\"name,omitempty\" bson:\"name\"`")
	if r != "name" {
		t.Errorf("expeted `name` but got: %s", r)
	}
}

func TestReadStructDirective1(t *testing.T) {
	errorFound = nil
	collFieldTypes = make(map[string]string)
	files := initCheckerSingleFile("sample3.go", "seeddata", src3)
	for _, f := range files {
		ast.Walk(&printASTVisitor{&info}, f)
		//the logging only happens on a fail test
		for k, v := range collFieldTypes {
			t.Logf("%s: %s\n", k, v)
		}
		if len(collFieldTypes) != 4 {
			t.Errorf("found: %d, expected: %d mongodb fields detected", len(collFieldTypes), 4)
		}
		if ret := collFieldTypes["\"xyz_company\".\"name\""]; ret != "string" {
			t.Errorf("found %q instead of \"string\"", ret)
		}
	}
}

func TestReadStructDirective2(t *testing.T) {
	errorFound = nil
	collFieldTypes = make(map[string]string)
	files := initCheckerSingleFile("sample4.go", "seeddata", src4)
	for _, f := range files {
		ast.Walk(&printASTVisitor{&info}, f)
		//the logging only happens on a fail test
		for k, v := range collFieldTypes {
			t.Logf("%s: %s\n", k, v)
		}
		if len(collFieldTypes) != 3 {
			t.Errorf("found: %d, expected: %d mongodb fields detected", len(collFieldTypes), 3)
		}
		if ret := collFieldTypes["\"xyz_company\".\"name\""]; ret != "string" {
			t.Errorf("wrong name field type. Found %q instead of \"string\"", ret)
		}
		if ret := collFieldTypes["\"xyz_company\".\"street\""]; ret != "string" {
			t.Errorf("wrong street field type. Found %q instead of \"string\"", ret)
		}
	}
}
func TestReadStructDirectiveTag(t *testing.T) {
	t.SkipNow()
	errorFound = nil
	collFieldTypes = make(map[string]string)
	files := initCheckerSingleFile("sample5.go", "seeddata", src5)
	for _, f := range files {
		ast.Walk(&printASTVisitor{&info}, f)
		//the logging only happens on a fail test
		for k, v := range collFieldTypes {
			t.Logf("%s: %s\n", k, v)
		}
		if len(collFieldTypes) != 3 {
			t.Errorf("found: %d, expected: %d mongodb fields detected", len(collFieldTypes), 3)
		}
		if ret := collFieldTypes["\"xyz_company\".\"zip_code\""]; ret != "string" {
			t.Errorf("wrong zip_code field type. Found %q instead of \"string\"", ret)
		}
		if ret := collFieldTypes["\"xyz_company\".\"name\""]; ret != "string" {
			t.Errorf("wrong name field type. Found %q instead of \"string\"", ret)
		}
		if ret := collFieldTypes["\"xyz_company\".\"street\""]; ret != "string" {
			t.Errorf("wrong street field type. Found %q instead of \"string\"", ret)
		}
	}
}

func TestExpectStrGotIntMultiItemMap1(t *testing.T) {
	errorFound = nil
	collFieldTypes = make(map[string]string)
	files := initCheckerSingleFile("sample6.go", "seeddata", src6)
	for _, f := range files {
		ast.Walk(&printASTVisitor{&info}, f)
		if errorFound.Actual != "int" {
			t.Errorf("actual: %s, expected: %s", errorFound.Actual, errorFound.Expected)
		}
	}
}

func TestExpectStrGotIntMultiItemMap2(t *testing.T) {
	errorFound = nil
	collFieldTypes = make(map[string]string)
	files := initCheckerSingleFile("sample7.go", "seeddata", src7)
	for _, f := range files {
		ast.Walk(&printASTVisitor{&info}, f)
		if errorFound.Expected != "" {
			t.Errorf("actual: %s, expected: %+v", errorFound.Actual, errorFound)
		}
	}
}
