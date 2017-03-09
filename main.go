package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io"
	"os"
	//"path/filepath"
	"strings"
)

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

func main() {
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
