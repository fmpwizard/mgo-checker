package main

import (
	"go/ast"
	"go/token"
	"go/types"
)

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
	files := initChecker()
	for _, f := range files {
		ast.Walk(&printASTVisitor{&info}, f)
	}
	typeReport()
}
