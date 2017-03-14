package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

var errorFound *ErrTypeInfo

//This will need a mutex
var collectionsMap = make(map[string]string)
var collectionsVarToNameMap = make(map[string]string)
var currentKey = ""

//var structTocollection = make(map[string]string)

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

	dirPath := flag.String("dir-path", "", "specify the path to the folder with go files to check")
	debug := flag.Bool("debug", false, "print extra debug information")
	flag.Parse()

	files := initChecker(*dirPath)
	for _, f := range files {
		ast.Walk(&printASTVisitor{&info}, f)
		if errorFound != nil {
			fmt.Println(errorFound)
		}
	}
	if *debug {
		typeReport()
	}
}
