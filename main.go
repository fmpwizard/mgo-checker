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
			if info.ObjectOf(n) != nil && info.ObjectOf(n).Type().String() == "*gopkg.in/mgo.v2.Collection" {
				details(n)
			}
		case *ast.CallExpr:
			if info.TypeOf(n.Fun).String() == "func(name string) *gopkg.in/mgo.v2.Collection" {
				//details(n.Fun)
				for _, arg := range n.Args {
					details(arg)
				}
			}
			fmt.Println()
		case *ast.SelectorExpr:
			if info.ObjectOf(n.Sel).Type().String() == "func(query interface{}) *gopkg.in/mgo.v2.Query" {
				details(n)
			}
			/*
				case *ast.SelectorExpr:
					if info.ObjectOf(n.Sel).Type().String() == "func(name string) *gopkg.in/mgo.v2.Collection" {
						fmt.Println("wwwwwwwwwwwwwwwwwwwwwwwww")
						details(n)
					}
			*/
		/*
			case *ast.AssignStmt:
					for _, x := range n.Lhs {
						fmt.Println("\ngoing in the left ")
						fmt.Printf("type is ========================: %+v\n", info.TypeOf(x.(ast.Expr)))
						if info.TypeOf(x.(ast.Expr)).String() == "*gopkg.in/mgo.v2.Collection" {
							details(x)
						}
					}

					for _, x := range n.Rhs {
						fmt.Println("\ngoing in the right ")
						if info.TypeOf(x.(ast.Expr)).String() == "*gopkg.in/mgo.v2.Collection" {
							details(x)
						}
					}
		*/
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
	fmt.Println("\n--------------------------------------------------")
	if node != nil {
		pos := fset.Position(node.Pos())
		fmt.Printf("\nThis is is!!1!!!!!!!!!!!!!!!!!!! %s: %s\n", pos, reflect.TypeOf(node).String())

		switch n := node.(type) {

		case *ast.BasicLit:
			fmt.Printf("\rhs: %+v\n", n.Value)
			if currentKey != "" {
				collectionsVarToNameMap[currentKey] = n.Value
				currentKey = ""
			}
		case *ast.SelectorExpr:
			details(n.Sel)
		case *ast.Ident:
			fmt.Printf("ident Type: ========================: %+v\n", info.ObjectOf(n).Type().String())
			fmt.Printf("ident Id: ========================: %+v\n", info.ObjectOf(n).Id()) //NAme and Id are the same

			//fmt.Printf("ident String: ========================: %+v\n", info.ObjectOf(n).String())
			/*
				if info.ObjectOf(n).Parent() != nil {
					for _, x := range info.ObjectOf(n).Parent().Names() {
						//fmt.Printf("ident Parent Name: ========================: %+v\n", x)
						fmt.Printf("ident Parent Lookup => Type.String(): ========================: %+v\n", info.ObjectOf(n).Parent().Lookup(x).Type().String())
						fmt.Printf("ident Parent Lookup => Id(): ========================: %+v\n", info.ObjectOf(n).Parent().Lookup(x).Id())
					}
					fmt.Printf("ident Parent String: ========================: %+v\n", info.ObjectOf(n).Parent().String())
				}
			*/
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
