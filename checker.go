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
			"%s:%d:%d: wrong mongodb field name, could not find field: %q",
			e.Filename,
			e.Line,
			e.Column,
			e.Actual,
		)
	}
	return fmt.Sprintf(
		"%s:%d:%d: wrong 'value' type, expected %q but got %q",
		e.Filename,
		e.Line,
		e.Column,
		e.Expected,
		e.Actual,
	)
}

func initChecker(dirPath string) []*ast.File {
	fset = token.NewFileSet()
	pcks, err := parser.ParseDir(fset, dirPath, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("failed to parse file ", err)
		os.Exit(1)
	}

	conf := &types.Config{
		Error: func(e error) {
			fmt.Println(e)
		},
		Importer: importer.Default(),
	}
	var files []*ast.File
	for pckPath, row := range pcks {
		for _, file := range row.Files {
			files = append(files, file)
		}
		pckg, err = conf.Check(pckPath, fset, files, &info)
		if err != nil {
			fmt.Printf("unexpected error: %v", err)
			os.Exit(1)
		}
	}
	return files
}

func initCheckerSingleFile(filePath, pkg string, src interface{}) []*ast.File {
	fset = token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	files := []*ast.File{f}
	if err != nil {
		fmt.Println("failed to parse file ", err)
		os.Exit(1)
	}

	conf := &types.Config{
		Error: func(e error) {
			fmt.Println(e)
		},
		Importer: importer.Default(),
	}

	pckg, err = conf.Check(pkg, fset, files, &info)
	if err != nil {
		fmt.Printf("unexpected error: %v", err)
		os.Exit(1)
	}

	return files
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
					if ret := getQueryFieldsInfo(arg); ret != nil {
						return nil
					}
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
	if errorFound != nil {
		//break away from Walk
		return nil
	}
	return v
}

func getQueryFieldsInfo(node ast.Node) *ErrTypeInfo {
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
					errorFound = &ErrTypeInfo{
						Expected: "",
						Actual:   collFieldTypeKey,
						Filename: pos.Filename,
						Column:   pos.Column,
						Line:     pos.Line,
					}
					return errorFound

				}
				if expectedType != actualType {
					errorFound = &ErrTypeInfo{
						Expected: expectedType,
						Actual:   actualType,
						Filename: pos.Filename,
						Column:   pos.Column,
						Line:     pos.Line,
					}
					return errorFound
				}
			}

		case *ast.CompositeLit:
			for _, row := range n.Elts {
				if ret := getQueryFieldsInfo(row); ret != nil {
					return ret
				}
			}
		}
	}
	return nil
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
