package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
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

//This will need a mutex
var collectionsMap = make(map[string]string)
var collectionsVarToNameMap = make(map[string]string)
var currentKey = ""

// "collection.field = string | int | bson.ObjectId"
var collFieldTypes = make(map[string]string)

var foundQ = false
var foundM = false

var fset *token.FileSet
var info = types.Info{
	Types: make(map[ast.Expr]types.TypeAndValue),
	Defs:  make(map[*ast.Ident]types.Object),
	Uses:  make(map[*ast.Ident]types.Object),
}

func main() {

	//fakenews:
	collFieldTypes["xyz_company.name"] = "string"
	collFieldTypes["xyz_company.zip_code"] = "string"

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

	_, err = conf.Check("seeddata", fset, []*ast.File{f}, &info)
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
	//fmt.Println("--------------------------------------------------")
	if node != nil {
		pos := fset.Position(node.Pos())
		fmt.Printf("%s: %s", pos, reflect.TypeOf(node).String())
		switch n := node.(type) {
		case *ast.Ident:
			obj := info.ObjectOf(n)
			if obj != nil {
				switch obj.Type().String() {
				case "*gopkg.in/mgo.v2.Collection", "func(query interface{}) *gopkg.in/mgo.v2.Query":
					details(n)
				default:
					fmt.Println("wwwwwwwwwwww ", info.ObjectOf(n).Type().String())
				}
			}

		case *ast.CallExpr:
			switch info.TypeOf(n.Fun).String() {
			case "func(name string) *gopkg.in/mgo.v2.Collection", "func(query interface{}) *gopkg.in/mgo.v2.Query":

				for _, arg := range n.Args {
					details(arg)
				}
			default:
				fmt.Println("\n\nast.CallExpr ================> ", info.TypeOf(n.Fun).String())
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

func details(node ast.Node) {
	//fmt.Println("\n--------------------------------------------------")
	if node != nil {
		pos := fset.Position(node.Pos())
		fmt.Printf("\nThis is is!!1!!!!!!!!!!!!!!!!!!! %s: %s\n", pos, reflect.TypeOf(node).String())

		switch n := node.(type) {
		case *ast.KeyValueExpr:
			fmt.Printf("value type: %s\n", info.TypeOf(n.Value))
			fmt.Printf("key fields: %+v\n", n.Key)
		case *ast.CompositeLit:
			if info.TypeOf(n.Type) != nil {
				fmt.Println("type ", info.TypeOf(n.Type).String())
			}
			for _, row := range n.Elts {
				fmt.Printf("elt %+v\n", row)
				details(row)
				if info.TypeOf(row) != nil {
					fmt.Println("row ", info.TypeOf(row).String())
				}
			}
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
}
