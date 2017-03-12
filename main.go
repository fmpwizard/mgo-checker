package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"reflect"
)

var src = `
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

// ErrTypeInfo holds information about the incorrect type parameter found.
type ErrTypeInfo struct {
	Expected string
	Actual   string
	Filename string
	Line     int
	Column   int
}

func (e ErrTypeInfo) String() string {
	if e.Expected == "" {
		return fmt.Sprintf(
			"%s:%d:%d: wrong mongodb field name, could not find field: %s",
			e.Filename,
			e.Line,
			e.Column,
			e.Actual,
		)
	}
	return fmt.Sprintf(
		"%s:%d:%d: wrong 'value' type, expected %s but got %s",
		e.Filename,
		e.Line,
		e.Column,
		e.Expected,
		e.Actual,
	)
}

var errorFound ErrTypeInfo

//This will need a mutex
var collectionsMap = make(map[string]string)
var collectionsVarToNameMap = make(map[string]string)
var currentKey = ""

// "collection.field = string | int | bson.ObjectId"
var collFieldTypes = make(map[string]string)

var fset *token.FileSet
var pckg *types.Package
var info = types.Info{
	Types:      make(map[ast.Expr]types.TypeAndValue),
	Defs:       make(map[*ast.Ident]types.Object),
	Uses:       make(map[*ast.Ident]types.Object),
	Implicits:  make(map[ast.Node]types.Object),
	Selections: make(map[*ast.SelectorExpr]*types.Selection),
	Scopes:     make(map[ast.Node]*types.Scope),
}

func main() {

	//fakenews:
	collFieldTypes[`"xyz_company"."name"`] = "string"
	collFieldTypes[`"xyz_company"."zip_code"`] = "string"

	fset = token.NewFileSet()

	f, err := parser.ParseFile(fset, "./seeddata/sample.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println("failed to parse file ", err)
	}

	conf := &types.Config{
		Error: func(e error) {
			fmt.Println(e)
		},
		Importer: importer.Default(),
	}

	pckg, err = conf.Check("seeddata", fset, []*ast.File{f}, &info)
	if err != nil {
		fmt.Printf("unexpected error: %v", err)
	}

	fmt.Println("len ", len(info.Types))
	ast.Walk(&printASTVisitor{&info}, f)
	typeReport()
}

type printASTVisitor struct {
	info *types.Info
}

func (v *printASTVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		pos := fset.Position(node.Pos())
		fmt.Printf("%s: %s", pos, reflect.TypeOf(node).String())
		switch n := node.(type) {

		case *ast.Ident:
			obj := info.ObjectOf(n)
			if obj != nil {
				switch obj.Type().String() {
				case "*gopkg.in/mgo.v2.Collection":
					details(n)
				default:
					//fmt.Println("wwwwwwwwwwww ", info.ObjectOf(n).Type().String())
				}
			}

		case *ast.CallExpr:
			switch info.TypeOf(n.Fun).String() {
			case "func(name string) *gopkg.in/mgo.v2.Collection":
				for _, arg := range n.Args {
					details(arg)
				}
			case "func(query interface{}) *gopkg.in/mgo.v2.Query":
				for _, arg := range n.Args {
					getQueryFieldsInfo(arg)
				}
			default:
				//fmt.Println("\n\nast.CallExpr ================> ", info.TypeOf(n.Fun).String())
			}
			fmt.Println()

		case ast.Expr:
			t := v.info.TypeOf(node.(ast.Expr))
			if t != nil {
				fmt.Printf(" : %s", t.String())
			}
		}
		fmt.Println()
	}
	return v
}

func getQueryFieldsInfo(node ast.Node) {
	if node != nil {
		pos := fset.Position(node.Pos())
		fmt.Printf("\n%s: %s\n", pos, reflect.TypeOf(node).String())

		switch n := node.(type) {
		case *ast.KeyValueExpr:
			varName := pckg.Scope().Innermost(node.Pos()).Lookup("testCollection").Id()
			if collectionName, ok := collectionsVarToNameMap[varName]; ok {
				collFieldTypeKey := collectionName + "." + n.Key.(*ast.BasicLit).Value
				expectedType, ok := collFieldTypes[collFieldTypeKey]
				actualType := info.TypeOf(n.Value).String()
				pos := fset.Position(n.Key.Pos())
				if !ok {
					errorFound = ErrTypeInfo{
						Expected: "",
						Actual:   collFieldTypeKey,
						Filename: pos.Filename,
						Column:   pos.Column,
						Line:     pos.Line,
					}
					fmt.Println(errorFound)
					os.Exit(0)
				}
				if expectedType != actualType {
					errorFound = ErrTypeInfo{
						Expected: expectedType,
						Actual:   collFieldTypeKey,
						Filename: pos.Filename,
						Column:   pos.Column,
						Line:     pos.Line,
					}
					fmt.Println(errorFound)
					os.Exit(0)
				}
			}

		case *ast.CompositeLit:
			for _, row := range n.Elts {
				getQueryFieldsInfo(row)
			}
		}
	}
}

func details(node ast.Node) {
	if node != nil {
		pos := fset.Position(node.Pos())
		fmt.Printf("\nThis is is!!1!!!!!!!!!!!!!!!!!!! %s: %s\n", pos, reflect.TypeOf(node).String())

		switch n := node.(type) {
		case *ast.BasicLit:
			fmt.Printf("\rBasicLit: %+v\n", n.Value)
			if currentKey != "" {
				collectionsVarToNameMap[currentKey] = n.Value
				currentKey = ""
			}

		case *ast.Ident:
			fmt.Printf("ident Type: ========================: %+v\n", info.ObjectOf(n).Type().String())
			fmt.Printf("ident Id: ========================: %+v\n", info.ObjectOf(n).Id()) //NAme and Id are the same

			if info.ObjectOf(n).Type().String() == "func(name string) *gopkg.in/mgo.v2.Collection" {
				fmt.Printf("found it!!!!!!!!!!!!!!!\n")
			}
			if n.Obj != nil {
				collectionsMap[info.ObjectOf(n).Id()] = info.ObjectOf(n).Type().String()
				currentKey = info.ObjectOf(n).Id()
			}
		}
	}
}

func typeReport() {
	fmt.Println("Found these variables:")
	for k, v := range collectionsMap {
		fmt.Printf("%s => %s\n", k, v)
	}
	fmt.Println("Found these collections:")
	for k, v := range collectionsVarToNameMap {
		fmt.Printf("%s => %s\n", k, v)
	}
	fmt.Println("MongoDB info:")
	for k, v := range collFieldTypes {
		fmt.Printf("%s => %s\n", k, v)
	}
}
