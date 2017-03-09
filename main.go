package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"strings"
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

// AppInfo holds all the models we collected
type AppInfo struct {
	Models []Model
}

// Model has the collection name and a slice of Fields. It also has the package where the struct was defined
type Model struct {
	Package string
	Name    string
	Fields  []Field
}

// Field holds the field's name and type
type Field struct {
	Name string
	Type string
}

var appInfo AppInfo

var kPath = "./seeddata"
var report = false
var fset *token.FileSet

func main() {
	fset = token.NewFileSet()

	/*
		pkgs, e := parser.ParseDir(fset, kPath, nil, parser.ParseComments)
		if e != nil {
			log.Fatal(e)
			return
		}

		astf := make([]*ast.File, 0)
		for _, pkg := range pkgs {

			fmt.Printf("package %v\n", pkg.Name)
			for fn, f := range pkg.Files {
				fmt.Printf("file %v\n", fn)
				astf = append(astf, f)
			}
		}
	*/

	/*
		config := &types.Config{
			Error: func(e error) {
				fmt.Println(e)
			},
			Importer: importer.Default(),
		}
	*/
	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	/*
		pkg, e := config.Check(kPath, fset, astf, &info)
		if e != nil {
			fmt.Println(e)
		}
		fmt.Printf("types.Config.Check got %v\n", pkg.String())
	*/
	//for _, pkg := range pkgs {
	//getDocs(pkg)
	//}

	f, err := parser.ParseFile(fset, "sample.go", src, parser.ParseComments)
	if err != nil {
		fmt.Println("failed to parse file ", err)
	}
	ast.Walk(&PrintASTVisitor{&info}, f)

	/*
		for _, f := range astf {
			ast.Walk(&PrintASTVisitor{&info}, f)
		}
	*/

}

func getDocs(pkg *ast.Package) {
	p := doc.New(pkg, "./", 0)
	for _, t := range p.Types {
		//fmt.Println("  type", t.Name)
		//fmt.Printf("    docs:\n%v", t.Doc)
		if isMgoStruct(t.Doc) {
			m := getMgoCollName(t.Doc)
			fmt.Printf("collection name: %s\n", m)
		}
	}
	fmt.Println("step ======================= step")

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

func main1() {
	//filepath.Walk("./seeddata", walk)
	//fmt.Println("=============")
	docs()
}

func docs() {
	fset := token.NewFileSet() // positions are relative to fset

	d, err := parser.ParseDir(fset, "./seeddata", nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, f := range d {
		//fmt.Println("package", k)
		p := doc.New(f, "./", 0)
		for _, t := range p.Types {
			//fmt.Println("  type", t.Name)
			//fmt.Printf("    docs:\n%v", t.Doc)
			if isMgoStruct(t.Doc) {
				m := getMgoCollName(t.Doc)
				fmt.Printf("collection name: %s\n", m)
			}
		}
		fmt.Println("step ======================= step")
		ast.Inspect(f, func(n ast.Node) bool {
			var s string
			switch x := n.(type) {
			case *ast.FuncLit:
				fmt.Println("xxxxxxxxxxxxxxxxxxxxxx")
				for _, v := range x.Body.List {
					fmt.Printf("fn: %+v\n", v)
				}
			case *ast.FuncDecl:
				fmt.Println("---------------------------")
				fmt.Printf("fn: %+v\n", x.Name)

			/*case *ast.BasicLit:
				//fmt.Println("a")
				s = x.Value
			case *ast.Ident:
				//fmt.Println("b")
				if x.Obj != nil {
					fmt.Println("c: ", x.Obj.Kind.String())
					switch d := x.Obj.Decl.(type) {
					case *ast.Field:
						fmt.Printf("field Tag: %s\n", d.Tag.Value)
						fmt.Printf("field Type: %+v\n", d.Type)
						for _, n := range d.Names {
							fmt.Printf("field Name: %+v\n", n)
						}
					}
				}
				s = x.Name
			*/
			default:
				fmt.Printf("00000000000000000000000 %T\n", x)

			}
			if s != "" {
				fmt.Printf("%s:\t%s\n", fset.Position(n.Pos()), s)
			}
			return true
		})

	}
}

func walk(path string, info os.FileInfo, err error) error {
	fmt.Println("reading: ", path)
	if strings.HasSuffix(info.Name(), ".go") {
		findDirective(path)
	}
	return err
}

func findDirective(fpath string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	input := bufio.NewReader(f)
	for {
		var buf []byte
		buf, err = input.ReadSlice('\n')
		if err == bufio.ErrBufferFull {
			return bufio.ErrBufferFull
		}
		if err != nil {
			// Check for marker at EOF without final \n.
			if err == io.EOF && isMGoDirective(buf) {
				err = io.ErrUnexpectedEOF
				return err
			}
			break
		}
		if isMGoDirective(buf) {
			fmt.Println(string(buf))
		}
	}

	return nil
}

func isMgoStruct(in string) bool {
	return strings.Contains(in, "mgo:model")
}

func getMgoCollName(in string) string {
	x := strings.Split(in, ":")
	if len(x) == 3 {
		return strings.TrimSpace(x[2])
	}
	return ""
}

func isMGoDirective(in []byte) bool {
	return bytes.HasPrefix(in, []byte("//mgo:model"))
}
