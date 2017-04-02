package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"strings"
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

var errorFound *ErrTypeInfo

//This will need a mutex
var collectionsMap = make(map[string]string)
var collectionsVarToNameMap = make(map[string]string)
var currentKey = ""

//var structTocollection = make(map[string]string)

// "collection.field = string | int | bson.ObjectId"
var collFieldTypes = make(map[string]string)

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
	//debug := flag.Bool("debug", false, "print extra debug information")
	flag.Parse()
	files, err := ioutil.ReadDir(*dirPath)
	if err != nil {
		fmt.Println("faile to read dir ", err)
		os.Exit(1)
	}
	doPackage(*dirPath, files)
	/*
		files, fset, _ := initChecker(*dirPath)
		for _, f := range files {
			visitor := &printASTVisitor{
				info: &info,
				fset: fset,
			}
			ast.Walk(visitor, f)
			if errorFound != nil {
				fmt.Println(errorFound)
			}
		}
		if *debug {
			typeReport()
		}
	*/
}

// doPackage analyzes the single package constructed from the named files.
// It returns the parsed Package or nil if none of the files have been checked.
func doPackage(basePath string, names []os.FileInfo) *Package {

	var files []*File
	var astFiles []*ast.File
	var funcUsingCollection = make(map[string]string, 0)
	fs := token.NewFileSet()
	for _, name := range names {
		data, err := ioutil.ReadFile(basePath + "/" + name.Name())
		if err != nil {
			// Warn but continue to next package.
			fmt.Printf("1 %s: %s", basePath+"/"+name.Name(), err)
			return nil
		}
		var parsedFile *ast.File
		if strings.HasSuffix(name.Name(), ".go") {
			parsedFile, err = parser.ParseFile(fs, name.Name(), data, parser.ParseComments)
			if err != nil {
				fmt.Printf("%s: %s", name, err)
				return nil
			}
			astFiles = append(astFiles, parsedFile)
		}
		files = append(files, &File{
			fset:                fs,
			content:             data,
			name:                name.Name(),
			file:                parsedFile,
			funcUsingCollection: funcUsingCollection,
		})
	}
	if len(astFiles) == 0 {
		return nil
	}
	pkg := new(Package)
	pkg.path = astFiles[0].Name.Name
	pkg.files = files
	pkg.types = info.Types
	pkg.defs = info.Defs
	pkg.selectors = info.Selections
	pkg.uses = info.Uses
	// Type check the package.
	conf := &types.Config{
		Error: func(e error) {
			fmt.Println(e)
		},
		Importer: importer.Default(),
	}
	_, err := conf.Check(pkg.path, fs, astFiles, &info)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	// Check.
	/*
		chk := make(map[ast.Node][]func(*File, ast.Node))
		for typ, set := range checkers {
			for name, fn := range set {
				if vet(name) {
					chk[typ] = append(chk[typ], fn)
				}
			}
		}
	*/
	for _, file := range files {
		file.pkg = pkg
		//file.basePkg = basePkg
		//file.checkers = chk
		if file.file != nil {
			file.walkFile(file.name, file.file)
		}
	}
	fmt.Println("results:")
	for k, v := range collFieldTypes {
		fmt.Printf("k: %s, v: %s\n", k, v)
	}

	return pkg
}
