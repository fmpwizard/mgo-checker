package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
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

var report = false
var fset *token.FileSet

func main() {
	fset = token.NewFileSet()

	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	f, err := parser.ParseFile(fset, "sample.go", src, parser.ParseComments)
	if err != nil {
		fmt.Println("failed to parse file ", err)
	}
	ast.Walk(&PrintASTVisitor{&info}, f)
}

type PrintASTVisitor struct {
	info *types.Info
}

func (v *PrintASTVisitor) Visit(node ast.Node) ast.Visitor {
	fmt.Println("++++++++++++++++++++++++++++++++")
	if node != nil {
		//fmt.Printf("%s", reflect.TypeOf(node).String())

		switch n := node.(type) {
		case *ast.CommentGroup:
			fmt.Printf("comment: %+v\n", n.Text())
		case *ast.Ident:
			//fmt.Printf(" : %+v", fset.Position(node.Pos()))
			//fmt.Printf("-------------------------------here: %+v\n", n.Name)
			t := v.info.TypeOf(node.(ast.Expr))
			fmt.Printf("--------------------- Name : %s\n", n.Name)
			if t != nil {
				if t.String() == "*gopkg.in/mgo.v2.Collection" {
					fmt.Printf(" : %+v", fset.Position(node.Pos()))
					fmt.Printf("---------------------Found collections! : %s\n", t.String())
					fmt.Printf("---------------------node fields: : %+v\n", n)
					report = true
				}
				//report = false
			}
			//report = false

		case ast.Expr:
			if report {
				fmt.Printf("%+v", fset.Position(node.Pos()))
				t := v.info.TypeOf(node.(ast.Expr))
				if t != nil {
					fmt.Printf(" Expr: %s\n", t.String())
				}
			} else {
				fmt.Printf("report: %v, type: %+v\n", report, n)
			}
		default:
			fmt.Printf("report: %v, type: %+v\n", report, n)
		}
	}
	return v
}
