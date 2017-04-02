package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"strings"
)

var (
	httpResponseType types.Type
	httpClientType   types.Type
	errorType        *types.Interface
	exitCode         = 0

	// Each of these vars has a corresponding case in (*File).Visit.
	assignStmt    *ast.AssignStmt
	binaryExpr    *ast.BinaryExpr
	callExpr      *ast.CallExpr
	compositeLit  *ast.CompositeLit
	exprStmt      *ast.ExprStmt
	funcDecl      *ast.FuncDecl
	funcLit       *ast.FuncLit
	genDecl       *ast.GenDecl
	interfaceType *ast.InterfaceType
	rangeStmt     *ast.RangeStmt
	returnStmt    *ast.ReturnStmt
	structType    *ast.StructType
)

func init() {
	errorType = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)

}

func getVarAndCollectionName(f *File, node ast.Node) {
	//diego
	fmt.Println("ssssssssss")
	call := node.(*ast.CallExpr)
	if !isFuncC(f, call) {
		return // the function call is not related to this check.
	}

	finder := &blockStmtFinder{node: call}
	ast.Walk(finder, f.file)
	stmts := finder.stmts()
	asg, ok := stmts[0].(*ast.AssignStmt)
	if !ok {
		return // the first statement is not assignment.
	}

	w, ok := stmts[1].(*ast.ExprStmt)
	if !ok {
		return
	}
	fmt.Printf("here %+v\n", w.X)
	collectionVar := rootIdent(asg.Lhs[0])
	fmt.Printf("Boom %+v\n", collectionVar)
	if collectionVar == nil {
		return // could not find the http.Response in the assignment.
	}

	collName, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return
	}
	fmt.Printf("row : %+v\n", collName.Value)
	//DIego : continue here, run
	// go install && mgo-checker -dir-path=seeddata
	// and see how to best store the var => colletion name values/
	// to account for files with same var name but diff scopes

	for _, row := range stmts {
		usage, ok := row.(*ast.ExprStmt)
		if ok {
			if f.pkg.types[usage.X.(*ast.CallExpr).Args[0]].Type.String() == "*gopkg.in/mgo.v2.Collection" {
				x, ok := usage.X.(*ast.CallExpr).Fun.(*ast.Ident)
				if ok {
					//useful info for a line like:
					//findByZip(companyColl, "diego")
					//x.Obj.Decl.(*ast.FuncDecl).Type.Params.List[0].Type
					fmt.Printf("============ %+v\n", x.Name)
					f.funcUsingCollection[f.pkg.path+"."+x.Name] = collName.Value
				}
			}
		}
	}
}

// isFuncC checks whether the given call expression is on
// either a function of the net/http package or a method of http.Client that
// returns (*http.Response, error).
func isFuncC(f *File, expr *ast.CallExpr) bool {
	fun, _ := expr.Fun.(*ast.SelectorExpr)
	sig, _ := f.pkg.types[fun].Type.(*types.Signature)
	fmt.Println("dddddddddd ", fun)
	fmt.Println("dddddddddd ", sig)

	if sig == nil {
		return false // the call is not on of the form x.f()
	}

	res := sig.Results()

	fmt.Println("1111 ", res.At(0).Type().String()) //*gopkg.in/mgo.v2.Collection when it is good
	if res.At(0).Type().String() == "*gopkg.in/mgo.v2.Collection" {
		fmt.Println("found C func")
		return true
	}
	if res.Len() != 2 {
		return false // the function called does not return two values.
	}
	if ptr, ok := res.At(0).Type().(*types.Pointer); !ok || !types.Identical(ptr.Elem(), httpResponseType) {
		return false // the first return type is not *http.Response.
	}
	if !types.Identical(res.At(1).Type().Underlying(), errorType) {
		return false // the second return type is not error
	}

	typ := f.pkg.types[fun.X].Type
	if typ == nil {
		id, ok := fun.X.(*ast.Ident)
		return ok && id.Name == "http" // function in net/http package.
	}

	if types.Identical(typ, httpClientType) {
		return true // method on http.Client.
	}
	ptr, ok := typ.(*types.Pointer)
	return ok && types.Identical(ptr.Elem(), httpClientType) // method on *http.Client.
}

// blockStmtFinder is an ast.Visitor that given any ast node can find the
// statement containing it and its succeeding statements in the same block.
type blockStmtFinder struct {
	node  ast.Node       // target of search
	stmt  ast.Stmt       // innermost statement enclosing argument to Visit
	block *ast.BlockStmt // innermost block enclosing argument to Visit.
}

// Visit finds f.node performing a search down the ast tree.
// It keeps the last block statement and statement seen for later use.
func (f *blockStmtFinder) Visit(node ast.Node) ast.Visitor {
	if node == nil || f.node.Pos() < node.Pos() || f.node.End() > node.End() {
		return nil // not here
	}
	switch n := node.(type) {
	case *ast.BlockStmt:
		f.block = n
	case ast.Stmt:
		f.stmt = n
	}
	if f.node.Pos() == node.Pos() && f.node.End() == node.End() {
		return nil // found
	}
	return f // keep looking
}

// stmts returns the statements of f.block starting from the one including f.node.
func (f *blockStmtFinder) stmts() []ast.Stmt {
	for i, v := range f.block.List {
		if f.stmt == v {
			return f.block.List[i:]
		}
	}
	return nil
}

// rootIdent finds the root identifier x in a chain of selections x.y.z, or nil if not found.
func rootIdent(n ast.Node) *ast.Ident {
	switch n := n.(type) {
	case *ast.SelectorExpr:
		return rootIdent(n.X)
	case *ast.Ident:
		return n
	default:
		return nil
	}
}

// File is a wrapper for the state of a file used in the parser.
// The parse tree walkers are all methods of this type.
type File struct {
	pkg     *Package
	fset    *token.FileSet
	name    string
	content []byte
	file    *ast.File

	//funcUsingCollection map[ast.Expr]string //map of line like findByZip(companyColl, "diego") => xyz_company  which is the colletion name
	funcUsingCollection map[string]string //map of line like seeddata.findByZip: xyz_company  which is the package.funcName: colletion name
	//b       bytes.Buffer // for use by methods

	// Parsed package "foo" when checking package "foo_test"
	//basePkg *Package

	// The objects that are receivers of a "String() string" method.
	// This is used by the recursiveStringer method in print.go.
	//stringers map[*ast.Object]bool

	// Registered checkers to run.
	//checkers map[ast.Node][]func(*File, ast.Node)
}

type Package struct {
	path      string
	defs      map[*ast.Ident]types.Object
	uses      map[*ast.Ident]types.Object
	selectors map[*ast.SelectorExpr]*types.Selection
	types     map[ast.Expr]types.TypeAndValue
	spans     map[types.Object]Span
	files     []*File
	typesPkg  *types.Package
}

// Span stores the minimum range of byte positions in the file in which a
// given variable (types.Object) is mentioned. It is lexically defined: it spans
// from the beginning of its first mention to the end of its last mention.
// A variable is considered shadowed (if *strictShadowing is off) only if the
// shadowing variable is declared within the span of the shadowed variable.
// In other words, if a variable is shadowed but not used after the shadowed
// variable is declared, it is inconsequential and not worth complaining about.
// This simple check dramatically reduces the nuisance rate for the shadowing
// check, at least until something cleverer comes along.
//
// One wrinkle: A "naked return" is a silent use of a variable that the Span
// will not capture, but the compilers catch naked returns of shadowed
// variables so we don't need to.
//
// Cases this gets wrong (TODO):
// - If a for loop's continuation statement mentions a variable redeclared in
// the block, we should complain about it but don't.
// - A variable declared inside a function literal can falsely be identified
// as shadowing a variable in the outer function.
//
type Span struct {
	min token.Pos
	max token.Pos
}

// setExit sets the value for os.Exit when it is called, later. It
// remembers the highest value.
func setExit(err int) {
	if err > exitCode {
		exitCode = err
	}
}

// Bad reports an error and sets the exit code..
func (f *File) Bad(pos token.Pos, args ...interface{}) {
	f.Warn(pos, args...)
	setExit(1)
}

// Badf reports a formatted error and sets the exit code.
func (f *File) Badf(pos token.Pos, format string, args ...interface{}) {
	f.Warnf(pos, format, args...)
	setExit(1)
}

// loc returns a formatted representation of the position.
func (f *File) loc(pos token.Pos) string {
	if pos == token.NoPos {
		return ""
	}
	// Do not print columns. Because the pos often points to the start of an
	// expression instead of the inner part with the actual error, the
	// precision can mislead.
	posn := f.fset.Position(pos)
	return fmt.Sprintf("%s:%d", posn.Filename, posn.Line)
}

// locPrefix returns a formatted representation of the position for use as a line prefix.
func (f *File) locPrefix(pos token.Pos) string {
	if pos == token.NoPos {
		return ""
	}
	return fmt.Sprintf("%s: ", f.loc(pos))
}

// Warn reports an error but does not set the exit code.
func (f *File) Warn(pos token.Pos, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s%s", f.locPrefix(pos), fmt.Sprintln(args...))
}

// Warnf reports a formatted error but does not set the exit code.
func (f *File) Warnf(pos token.Pos, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s%s\n", f.locPrefix(pos), fmt.Sprintf(format, args...))
}

// walkFile walks the file's tree.
func (f *File) walkFile(name string, file *ast.File) {
	fmt.Println("Checking file", name)
	ast.Walk(f, file)
}

// Visit implements the ast.Visitor interface.
func (f *File) Visit(node ast.Node) ast.Visitor {
	//var key ast.Node
	switch node.(type) {
	/*
		case *ast.AssignStmt:
				key = assignStmt
			case *ast.BinaryExpr:
				key = binaryExpr
	*/
	case *ast.CallExpr:
		//key = callExpr
		getVarAndCollectionName(f, node)
		/*
			case *ast.CompositeLit:
				key = compositeLit
			case *ast.ExprStmt:
				key = exprStmt
		*/
	case *ast.FuncDecl:
		findFnUsingCollection(f, node)
		/*
			case *ast.FuncLit:
					key = funcLit
		*/
	case *ast.GenDecl:
		getFieldNameToTypeMap(f, node)
		/*
			case *ast.InterfaceType:
				key = interfaceType
			case *ast.RangeStmt:
				key = rangeStmt
			case *ast.ReturnStmt:
				key = returnStmt
			case *ast.StructType:
				key = structType
		*/
	}

	/*
		for _, fn := range f.checkers[key] {
			fn(f, node)
		}
	*/
	return f
}

func findFnUsingCollection(f *File, node ast.Node) {
	if n, ok := node.(*ast.FuncDecl); ok {
		if collName, ok := f.funcUsingCollection[f.pkg.path+"."+n.Name.Name]; ok {
			fmt.Printf("Found matching fn: %s using collection: %+v\n", n.Name.Name, collName)
			for _, stmt := range n.Body.List {
				//fmt.Printf("stmt: %+v\n", stmt)
				if exp, ok := stmt.(*ast.ExprStmt); ok {
					fmt.Printf("expStmt %+v\n", exp.X)
				}
				if exp, ok := stmt.(*ast.AssignStmt); ok {
					for _, assign := range exp.Rhs {
						if callExp, ok := assign.(*ast.CallExpr); ok {
							for _, v := range callExp.Args { //Diego continue here
								fmt.Printf(" ================>>>>>>>>>>>> arg: %+v\n", callExp.Fun)
								fmt.Printf(" ================>>>>>>>>>>>> arg: %+v\n", v)
							}
						}
					}
				}
				//ExprStmt or AssignStmt
			}
		}
	}
}

func getCollectionNameIThink(f *File, node ast.Node) {

	if n, ok := node.(*ast.Ident); ok {
		if n.Obj != nil {
			switch info.ObjectOf(n).Type().String() {
			case "*gopkg.in/mgo.v2.Collection":
				collectionsMap[info.ObjectOf(n).Id()] = info.ObjectOf(n).Type().String()
				currentKey = info.ObjectOf(n).Id()
				fmt.Println("here is 1: ", info.ObjectOf(n).Id())
				fmt.Println("here is 2: ", info.ObjectOf(n).Type().String())
			}
		}
	}
}

func getFieldNameToTypeMap(f *File, node ast.Node) {
	if n, ok := node.(*ast.GenDecl); ok {

		if ok, t := getMgoCollectionFromComment(n.Doc.Text()); ok {
			for _, row := range n.Specs {
				for _, field := range row.(*ast.TypeSpec).Type.(*ast.StructType).Fields.List {
					cleanFieldName := ""
					if field.Tag != nil {
						cleanFieldName = fieldFromTag(field.Tag.Value)
					}
					if cleanFieldName == "" {
						for _, name := range field.Names {
							cleanFieldName = strings.ToLower(name.Name)
							collFieldTypes[fmt.Sprintf("%q", t)+"."+fmt.Sprintf("%q", cleanFieldName)] = info.TypeOf(field.Type).String()
						}
					} else {
						collFieldTypes[fmt.Sprintf("%q", t)+"."+fmt.Sprintf("%q", cleanFieldName)] = info.TypeOf(field.Type).String()
					}
				}
			}
		}
	}
}

/*
// gofmt returns a string representation of the expression.
func (f *File) gofmt(x ast.Expr) string {
	f.b.Reset()
	printer.Fprint(&f.b, f.fset, x)
	return f.b.String()
}
*/

func getQueryFieldsInfo(node ast.Node) *ErrTypeInfo {

	/* Diego: fix this
	if node != nil {

		switch n := node.(type) {
		case *ast.KeyValueExpr:
			varName := "" // diego fix this  pckg.Scope().Innermost(node.Pos()).Lookup("testCollection").Id()
			if collectionName, ok := collectionsVarToNameMap[varName]; ok {
				collFieldTypeKey := collectionName + "." + n.Key.(*ast.BasicLit).Value
				expectedType := collFieldTypes[collFieldTypeKey]
				actualType := info.TypeOf(n.Value).String()
				//diego fix this

				pos := fset.Position(n.Key.Pos())

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
				fmt.Println("22222222222222222222")
				if ret := getQueryFieldsInfo(row); ret != nil {
					return ret
				}
			}
		}
	}
	*/
	return nil
}

func getMgoCollectionFromComment(s string) (bool, string) {
	if strings.Contains(s, "mgo:model:") {
		start := strings.Index(s, "mgo:model:")
		return true, strings.TrimSpace(strings.TrimPrefix(s[start:], "mgo:model:"))
	}
	return false, ""
}

func fieldFromTag(s string) string {
	if strings.Contains(s, "bson:\"") {
		start := strings.Index(s, "bson:\"") + 6 //6 = len(bson:")
		s = s[start:]
		end := strings.Index(s, "\"")
		tag := s[:end]
		fields := strings.Split(tag, ",")
		if len(fields[0]) == 0 {
			return ""
		}
		return strings.TrimPrefix(tag, "bson:\"")
	} else if !strings.Contains(s, ":") {
		fields := strings.Split(s, ",")
		return fields[0]
	}

	return ""
}
