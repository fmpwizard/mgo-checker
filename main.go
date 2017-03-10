package main

import (
	"fmt"
	"go/ast"
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

var fset *token.FileSet
var info = types.Info{
	Types: make(map[ast.Expr]types.TypeAndValue),
	Defs:  make(map[*ast.Ident]types.Object),
	Uses:  make(map[*ast.Ident]types.Object),
}

func main() {
	fset = token.NewFileSet()

	f, err := parser.ParseFile(fset, "./seeddata/sample.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println("failed to parse file ", err)
	}
	ast.Walk(&printASTVisitor{&info}, f)
}

type printASTVisitor struct {
	info *types.Info
}

func (v *printASTVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		pos := fset.Position(node.Pos())
		fmt.Printf("%s: %s", pos, reflect.TypeOf(node).String())
		switch n := node.(type) {
		case *ast.AssignStmt:
			for _, x := range n.Rhs {
				fmt.Println("\ngoing in the right ")
				details(x)
			}
			for _, x := range n.Lhs {
				fmt.Println("\ngoing in the left ")
				details(x)
			}
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
	if node != nil {
		pos := fset.Position(node.Pos())
		fmt.Printf("\nThis is is!!1!!!!!!!!!!!!!!!!!!! %s: %s\n", pos, reflect.TypeOf(node).String())

		switch n := node.(type) {
		case *ast.CallExpr:
			details(n.Fun)
			for _, arg := range n.Args {
				details(arg)
			}
			fmt.Println()
		case *ast.BasicLit:
			fmt.Printf("\rhs: %+v\n", n.Value)
		case *ast.SelectorExpr:
			details(n.Sel)
		case *ast.Ident:
			fmt.Printf("ident name: %s\n", n.Name)
			if n.Obj != nil {
				fmt.Printf("ident kind %+v\n", n.Obj.Kind)
				fmt.Printf("ident type %+v\n", n.Obj.Type)
				fmt.Printf("reflect type: %s\n", reflect.TypeOf(node).String())
			}
		}
	}
}
