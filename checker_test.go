package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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

func initTestWrappers(filename string, data []byte) ([]*File, []*ast.File, *token.FileSet) {
	var files []*File
	var astFiles []*ast.File
	names := []string{filename}
	var funcUsingCollection = make(map[string]string, 0)
	fs := token.NewFileSet()
	for _, name := range names {
		parsedFile, err := parser.ParseFile(fs, name, data, parser.ParseComments)
		if err != nil {
			fmt.Printf("%s: %s", name, err)
			return nil, nil, fs
		}
		astFiles = append(astFiles, parsedFile)

		files = append(files, &File{
			fset:                fs,
			content:             data,
			name:                name,
			file:                parsedFile,
			funcUsingCollection: funcUsingCollection,
		})
	}
	return files, astFiles, fs
}

func TestExpectStrGotInt(t *testing.T) {
	files, astFiles, fset := initTestWrappers("sample1.go", []byte(src1))
	ret := doPackage(files, astFiles, fset)
	for _, errorFound := range ret.errors {
		if errorFound.Actual != "int" {
			t.Errorf("actual: %s, expected: %s", errorFound.Actual, errorFound.Expected)
		}
	}
}

func TestWrongFieldName(t *testing.T) {
	files, astFiles, fset := initTestWrappers("sample2.go", []byte(src2))
	ret := doPackage(files, astFiles, fset)
	for _, errorFound := range ret.errors {
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
	files, astFiles, fset := initTestWrappers("sample3.go", []byte(src3))
	ret := doPackage(files, astFiles, fset)
	if len(ret.collFieldTypes) != 4 {
		t.Errorf("found: %d, expected: %d mongodb fields detected", len(ret.collFieldTypes), 4)
	}
	if r := ret.collFieldTypes["\"xyz_company\".\"name\""]; r != "string" {
		t.Errorf("found %q instead of \"string\"", r)
	}
}

func TestReadStructDirective2(t *testing.T) {
	files, astFiles, fset := initTestWrappers("sample4.go", []byte(src4))
	ret := doPackage(files, astFiles, fset)

	if len(ret.collFieldTypes) != 3 {
		t.Errorf("found: %d, expected: %d mongodb fields detected", len(ret.collFieldTypes), 3)
	}
	if r := ret.collFieldTypes["\"xyz_company\".\"name\""]; r != "string" {
		t.Errorf("wrong name field type. Found %q instead of \"string\"", r)
	}
	if r := ret.collFieldTypes["\"xyz_company\".\"street\""]; r != "string" {
		t.Errorf("wrong street field type. Found %q instead of \"string\"", r)
	}
}

func TestReadStructDirectiveTag(t *testing.T) {
	files, astFiles, fset := initTestWrappers("sample5.go", []byte(src5))
	r := doPackage(files, astFiles, fset)

	if len(r.collFieldTypes) != 3 {
		t.Errorf("found: %d, expected: %d mongodb fields detected", len(r.collFieldTypes), 3)
	}
	if ret := r.collFieldTypes["\"xyz_company\".\"zip_code\""]; ret != "string" {
		t.Errorf("wrong zip_code field type. Found %q instead of \"string\"", ret)
	}
	if ret := r.collFieldTypes["\"xyz_company\".\"name\""]; ret != "string" {
		t.Errorf("wrong name field type. Found %q instead of \"string\"", ret)
	}
	if ret := r.collFieldTypes["\"xyz_company\".\"street\""]; ret != "string" {
		t.Errorf("wrong street field type. Found %q instead of \"string\"", ret)
	}
}

func TestExpectStrGotIntMultiItemMap1(t *testing.T) {
	files, astFiles, fset := initTestWrappers("sample6.go", []byte(src6))
	r := doPackage(files, astFiles, fset)
	for _, errorFound := range r.errors {
		if errorFound.Actual != "int" {
			t.Errorf("actual: %s, expected: %s", errorFound.Actual, errorFound.Expected)
		}
	}
}

func TestExpectStrGotIntMultiItemMap2(t *testing.T) {
	files, astFiles, fset := initTestWrappers("sample7.go", []byte(src7))
	r := doPackage(files, astFiles, fset)
	for _, errorFound := range r.errors {
		if errorFound.Expected != "" {
			t.Errorf("actual: %s, expected: %+v", errorFound.Actual, errorFound.Expected)
		}
	}
}
