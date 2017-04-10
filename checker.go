package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"os"
	"reflect"
	"strings"

	"golang.org/x/tools/go/loader"
)

var (
	errorType *types.Interface
	exitCode  = 0

// Each of these vars has a corresponding case in (*File).Visit.
)

func init() {
	errorType = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)
}

func getVarAndCollectionName(f *File, node ast.Node) {
	call := node.(*ast.CallExpr)
	if !isFuncC(f, call) {
		return // the function call is not related to this check.
	}
	finder := &blockStmtFinder{node: call}
	ast.Walk(finder, f.currentFile)
	fmt.Printf("stmts %+v\n", finder.block)
	stmts := finder.stmts()
	if stmts == nil {
		return
	}
	for _, v := range stmts {
		fmt.Printf("POS : %+v\n", f.program.Fset.Position(v.Pos()))
		fmt.Printf("f %+v\n", reflect.TypeOf(v))
		fmt.Printf("row is : %+v\n", v)
	}
	asg, ok := stmts[0].(*ast.AssignStmt)
	if !ok {
		return // the first statement is not assignment.
	}
	collectionVar := rootIdent(asg.Lhs[0])
	fmt.Printf("Boom %+v\n", collectionVar)
	if collectionVar == nil {
		return // could not find the http.Response in the assignment.
	}

	collName, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return
	}
	fmt.Printf("Found collection name: %+v\n", collName.Value)
	// go install && mgo-checker -dir-path=seeddata
	// and see how to best store the var => colletion name values/
	// to account for files with same var name but diff scopes
	for _, row := range stmts {
		fmt.Printf("row0 : %+v\n", row)
		ret := detectWrongTypeForField(f, row, collName.Value)
		if ret != nil {
			f.errors = append(f.errors, ret)
		}
		usage, ok := row.(*ast.ExprStmt)
		if ok {
			tpe := types.NewPointer(importType("gopkg.in/mgo.v2", "Collection"))
			// using types.IdenticalIgnoreTags keeps returning false here
			if len(usage.X.(*ast.CallExpr).Args) == 0 {
				return
			}
			fmt.Printf("row1 : %+v\n", f.lPkg.Types[usage.X.(*ast.CallExpr).Args[0]].Type.String())
			fmt.Printf("row11 : %+v\n", tpe.String())
			if f.lPkg.Types[usage.X.(*ast.CallExpr).Args[0]].Type.String() == tpe.String() {
				fmt.Printf("					=========== %+v\n", usage.X.(*ast.CallExpr).Fun)
				x, ok := usage.X.(*ast.CallExpr).Fun.(*ast.Ident)
				if ok {
					fmt.Printf("row2 : %+v\n", collName.Value)
					//useful info for a line like:
					//findByZip(companyColl, "diego")
					//x.Obj.Decl.(*ast.FuncDecl).Type.Params.List[0].Type
					f.funcUsingCollection[f.lPkg.Pkg.Name()+"."+x.Name] = collName.Value
					fmt.Printf("Boom			%+v\n", f.funcUsingCollection)
				}
			}
		}
	}
}

// importType returns the type denoted by the qualified identifier
// path.name, and adds the respective package to the imports map
// as a side effect. In case of an error, importType returns nil.
func importType(path, name string) types.Type {
	pkg, err := importer.Default().Import(path)
	if err != nil {
		// This can happen if the package at path hasn't been compiled yet.
		return nil
	}
	if obj, ok := pkg.Scope().Lookup(name).(*types.TypeName); ok {
		return obj.Type()
	}
	return nil
}

// isFuncC checks whether the given call expression is on
// either a function of the net/http package or a method of http.Client that
// returns (*http.Response, error).
func isFuncC(f *File, expr *ast.CallExpr) bool {
	fun, _ := expr.Fun.(*ast.SelectorExpr)
	sig, _ := f.lPkg.Types[fun].Type.(*types.Signature)
	fmt.Println("dddddddddd fun: ", fun)
	fmt.Println("dddddddddd sig: ", sig)

	if sig == nil {
		return false // the call is not on of the form x.f()
	}

	res := sig.Results()
	if res.Len() != 1 {
		return false // the function called does not return one value.
	}
	tpe := types.NewPointer(importType("gopkg.in/mgo.v2", "Collection"))
	fmt.Println("1111 ", res.At(0).Type().String()) //*gopkg.in/mgo.v2.Collection when it is good
	if res.At(0).Type().String() == tpe.String() {
		fmt.Println("found C func")
		return true
	}
	return false
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
	fmt.Printf("========== diego: f.node: %+v\n", f.node)
	fmt.Printf("========== diego: node: %+v\n", node)
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
		fmt.Printf("1 Bad!!!!!!!!!!!!!!!! %+v\n", f.node)
		return nil // found
	}
	return f // keep looking
}

// stmts returns the statements of f.block starting from the one including f.node.
func (f *blockStmtFinder) stmts() []ast.Stmt {
	if f.block == nil {
		return nil
	}
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
	program *loader.Program
	lPkg    *loader.PackageInfo
	errors  []*ErrTypeInfo
	// "collection.field = string | int | bson.ObjectId"
	collFieldTypes map[string]string
	currentFile    *ast.File

	//funcUsingCollection map[ast.Expr]string //map of line like findByZip(companyColl, "diego") => xyz_company  which is the colletion name
	funcUsingCollection map[string]string //map of line like seeddata.findByZip: xyz_company  which is the package.funcName: colletion name
}

// Package holds information for the current Go package we are processing
type Package struct {
	path      string
	defs      map[*ast.Ident]types.Object
	uses      map[*ast.Ident]types.Object
	selectors map[*ast.SelectorExpr]*types.Selection
	types     map[ast.Expr]types.TypeAndValue
	//spans     map[types.Object]Span
	files  []*File
	errors []*ErrTypeInfo

	// "collection.field = string | int | bson.ObjectId"
	collFieldTypes map[string]string

	//typesPkg  *types.Package
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
	//min token.Pos
	//max token.Pos
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
	posn := f.program.Fset.Position(pos)
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
func (f *File) walkFile(name string) {
	fmt.Println("Checking file", name)
	for _, file := range f.lPkg.Files {
		f.currentFile = file
		ast.Walk(f, file)
	}
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
		err := findFnUsingCollection(f, node)
		if err != nil {
			f.errors = append(f.errors, err)
			return nil
		}
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

func findFnUsingCollection(f *File, node ast.Node) *ErrTypeInfo {
	if n, ok := node.(*ast.FuncDecl); ok {
		if collectionName, ok := f.funcUsingCollection[f.lPkg.Pkg.Name()+"."+n.Name.Name]; ok {
			//fmt.Printf("										Found matching fn: %s using collection: %+v\n", n.Name.Name, collectionName)
			for _, stmt := range n.Body.List {
				//fmt.Printf("stmt: %+v\n", stmt)
				if exp, ok := stmt.(*ast.ExprStmt); ok {
					fmt.Printf("expStmt %+v\n", exp.X)
				}

				ret := detectWrongTypeForField(f, stmt, collectionName)
				if ret != nil {
					return ret
				}

			}
		}
	}
	return nil
}
func detectWrongTypeForField(f *File, stmt ast.Stmt, collectionName string) *ErrTypeInfo {
	if exp, ok := stmt.(*ast.AssignStmt); ok {
		for _, assign := range exp.Rhs {
			ret := detectWrongTypeForFieldInsideCallExpr(f, assign, collectionName)
			if ret != nil {
				return ret
			}
		}
	} else if n, ok := stmt.(*ast.ExprStmt); ok {
		ret := detectWrongTypeForFieldInsideCallExpr(f, n.X, collectionName)
		if ret != nil {
			return ret
		}
	} else if n, ok := stmt.(*ast.ReturnStmt); ok {
		for _, r := range n.Results {
			ret := detectWrongTypeForFieldInsideCallExpr(f, r, collectionName)
			if ret != nil {
				return ret
			}
		}
	}
	return nil
}

func detectWrongTypeForFieldInsideCallExpr(f *File, assign ast.Expr, collectionName string) *ErrTypeInfo {
	if callExp, ok := assign.(*ast.CallExpr); ok {
		if fn, ok := callExp.Fun.(*ast.SelectorExpr); ok {
			fmt.Printf(" ================>>>>>>>>>>>> POS: %+v\n", f.program.Fset.Position(fn.Pos()))
			fmt.Printf(" ================>>>>>>>>>>>> Fun1: %+v\n", fn.X)
			if a, ok := fn.X.(*ast.CallExpr); ok {
				for _, b := range a.Args {
					fmt.Printf(" ================>>>>>>>>>>>> POS: %+v\n", f.program.Fset.Position(b.Pos()))
					fmt.Printf(" ================>>>>>>>>>>>> Fun1: %+v\n", b)
					fmt.Printf(" ================>>>>>>>>>>>> type: %+v\n", f.lPkg.Types[b].Type.String())

					if c, ok := b.(*ast.Ident); ok {
						//*github.com/ascendantcompliance/blotterizer/vendor/gopkg.in/mgo.v2/bson.M
						fmt.Printf(" ================>>>>>>>>>>>> inner : %+v\n", c.Obj.Decl)
						fmt.Printf(" ================>>>>>>>>>>>> reflect: %+v\n", reflect.TypeOf(c.Obj.Decl))
						if d, ok := c.Obj.Decl.(*ast.AssignStmt); ok {
							for _, r := range d.Rhs {
								fmt.Printf(" ================>>>>>>>>>>>> inner : %+v\n", r)
								fmt.Printf(" ================>>>>>>>>>>>> reflect: %+v\n", reflect.TypeOf(r))
								if e, ok := r.(*ast.CallExpr); ok {
									fmt.Printf(" ================>>>>>>>>>>>> reflect now: %+v\n", reflect.TypeOf(e.Fun))

									for _, r := range e.Args {
										fmt.Printf(" ================>>>>>>>>>>>> args : %+v\n", r)
									}
									if ff, ok := e.Fun.(*ast.SelectorExpr); ok {
										/*
											for k, v := range f.lPkg.Defs {
												fmt.Printf("key: %+v, value: %+v\n", k, v)
											}
										*/
										//fmt.Printf(" ================>>>>>>>>>>>> ff.Sel.Obj : %+v\n", ff.Sel.Obj) //Diego this is nil becaue it is declared on a diff package
										//fmt.Printf(" ================>>>>>>>>>>>> ff.Sel : %+v\n", ff.Sel)
										fmt.Printf(" ================>>>>>>>>>>>> ff : %+v\n", ff)
										fmt.Printf(" ================>>>>>>>>>>>> f.lPkg.Selections[ff].Obj().Pos() : %+v\n", f.lPkg.Selections[ff].Obj().Pos())
										pos := f.lPkg.Selections[ff].Obj().Pos()
										//name := f.lPkg.Selections[ff].Obj().Name()
										//position := f.program.Fset.Position(pos)
										fmt.Printf(" <<<<<<<<<<<<<<<<<<<<<<<>>>>>>>>>>>>  : %+v\n", f.program.Fset.File(pos))
										for k, r := range f.program.AllPackages {
											if r != nil {
												fmt.Printf(" diego1 : %+v\n", k.Name())
												for _, astFile := range r.Files {
													for _, v := range astFile.Decls {
														if a, ok := v.(*ast.FuncDecl); ok {
															if a.Body != nil {
																fmt.Printf(" diego21 :%+v,  %+v, %+v, %+v\n", v, a.Pos(), a.Body.List, f.lPkg.Selections[ff].Obj().Pos()) //diego
																for _, b := range a.Body.List {
																	if c, ok := b.(*ast.AssignStmt); ok {
																		for _, d := range c.Rhs {
																			fmt.Printf(" diegom :%+v,  %+v\n", v, d)                 //diego
																			fmt.Printf(" diegom :%+v,  %+v\n", v, reflect.TypeOf(d)) //diego
																			if err := checkTypeComLit(f, d, collectionName); err != nil {
																				fmt.Println("diegom ", err)
																				return err
																			}
																		}
																	}
																}
																//fmt.Printf(" diego2 : %+v, %+v, %+v\n", k, reflect.TypeOf(v), a.Body.List) //diego
															}
														}
													}
												}
											}
										}
										fmt.Printf(" ================>>>>>>>>>>>> f.lPkg.Type[ff.Sel] : %+v\n", f.lPkg.Types[ff])
										//find out how to scan several pacakages and merge the ast
										if ff.Sel.Obj != nil {
											fmt.Printf(" ================>>>>>>>>>>>> reflect: %+v\n", reflect.TypeOf(ff.Sel.Obj.Decl))
										}
									}
								}
							}
						}
					}

					if err := checkTypeComLit(f, b, collectionName); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func checkTypeComLit(f *File, b ast.Expr, collectionName string) *ErrTypeInfo {
	if c, ok := b.(*ast.CompositeLit); ok {
		for _, d := range c.Elts {
			if keyValue, ok := d.(*ast.KeyValueExpr); ok {
				k := collectionName
				actualType := ""
				if mongoFieldNameUsedInMapQuery, ok := keyValue.Key.(*ast.BasicLit); ok {
					k = k + "." + mongoFieldNameUsedInMapQuery.Value
					//if f.collFieldTypes[k] == "time.Time" {
					fmt.Printf("diegomm0 %+v\n", collectionName)
					fmt.Printf("diegomm1 %+v\n", keyValue.Value)
					fmt.Printf("diegomm2 %+v\n", reflect.TypeOf(keyValue.Value))
					//}
				}

				switch mongoFieldTypeUsedInMapQuery := keyValue.Value.(type) {
				case *ast.BasicLit:
					//value is a literal value, not a variable
					if v, ok := f.lPkg.Types[mongoFieldTypeUsedInMapQuery]; ok {
						fmt.Printf("diegom %+v\n", mongoFieldTypeUsedInMapQuery)
						actualType = v.Type.String()
					}
				case *ast.Ident:
					//valule is a variable
					if v, ok := f.lPkg.Types[mongoFieldTypeUsedInMapQuery]; ok {
						actualType = v.Type.String()
					}
				}

				expectedType := f.collFieldTypes[k]
				pos := f.program.Fset.Position(keyValue.Value.Pos())
				if expectedType != actualType {
					errorFound := &ErrTypeInfo{
						Expected: expectedType,
						Actual:   actualType,
						Filename: pos.Filename,
						Column:   pos.Column,
						Line:     pos.Line,
					}
					return errorFound
				}
			}
		}
	}
	return nil
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
							f.collFieldTypes[fmt.Sprintf("%q", t)+"."+fmt.Sprintf("%q", cleanFieldName)] = f.lPkg.TypeOf(field.Type).String()
						}
					} else {
						f.collFieldTypes[fmt.Sprintf("%q", t)+"."+fmt.Sprintf("%q", cleanFieldName)] = f.lPkg.TypeOf(field.Type).String()
					}
				}
			}
		}
	}
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

// TrimVendorPath tries to remove the vendor path from s
func TrimVendorPath(s string) string {
	if strings.Contains(s, "/vendor/") {
		if s[:1] == "*" {
			return "*" + strings.Split(s, "/vendor/")[1]
		}
		return strings.Split(s, "/vendor/")[1]
	}
	return s
}
