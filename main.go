package main

import (
	"flag"
	"fmt"
	"go/ast"
	//"go/importer"
	//"go/parser"
	//"go/token"
	"go/types"
	//"io/ioutil"
	"os"
	//"strings"

	"golang.org/x/tools/go/loader"
)

var conf loader.Config

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

var info = types.Info{
	Types:      make(map[ast.Expr]types.TypeAndValue),
	Defs:       make(map[*ast.Ident]types.Object),
	Uses:       make(map[*ast.Ident]types.Object),
	Implicits:  make(map[ast.Node]types.Object),
	Selections: make(map[*ast.SelectorExpr]*types.Selection),
	Scopes:     make(map[ast.Node]*types.Scope),
}

func main() {

	//dirPath := flag.String("dir-path", "", "specify the path to the folder with go files to check")
	//debug := flag.Bool("debug", false, "print extra debug information")
	flag.Parse()
	dirs := os.Args[1:]

	conf.FromArgs(dirs, false)

	program, err := conf.Load()
	if err != nil {
		fmt.Println("failed ", err)
		os.Exit(1)
	}

	collFieldTypes := make(map[string]string)

	for _, createdPackage := range program.Imported {
		v := &File{
			program:        program,
			lPkg:           createdPackage,
			collFieldTypes: collFieldTypes,
		}
		v.walkFile("testing")
		//walkPackage(program, createdPackage)
	}

	for _, createdPackage := range program.Created {
		v := &File{
			program: program,
			lPkg:    createdPackage,
		}
		v.walkFile("testing")
	}

	/*

		fs := token.NewFileSet()
		var files []*File
		var astFiles []*ast.File

		for _, dir := range dirs {
			dir = strings.TrimSuffix(dir, "/")
			filesInfo, err := ioutil.ReadDir(dir)
			var filePkgPathAndNames []string
			for _, v := range filesInfo {
				filePkgPathAndNames = append(filePkgPathAndNames, dir+"/"+v.Name())
			}
			if err != nil {
				fmt.Println("faile to read dir ", err)
				os.Exit(1)
			}

			//files, astFiles, fset := genASTFilesAndFileWrapper(*dirPath, fileNames, fs)
			_files, _astFiles, fset := genASTFilesAndFileWrapper(filePkgPathAndNames, fs)
			files = append(files, _files...)
			astFiles = append(astFiles, _astFiles...)
			fs = fset
		}
		ret := doPackage(files, astFiles, fs)
		for _, v := range ret.errors {
			fmt.Println(v)
		}
	*/
}

func walkPackage(program *loader.Program, info *loader.PackageInfo) {
	for _, file := range info.Files {
		v := &visitor{
			program,
			info,
		}
		ast.Walk(v, file)
	}
}

type visitor struct {
	program *loader.Program
	pkg     *loader.PackageInfo
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if stmt, isRange := node.(*ast.RangeStmt); isRange {
		if _, isMap := v.pkg.TypeOf(stmt.X).(*types.Map); isMap {
			fmt.Println(v.program.Fset.Position(node.Pos()))
		}
	}
	return v
}

/*
func genASTFilesAndFileWrapper(filePkgPathAndNames []string, fs *token.FileSet) ([]*File, []*ast.File, *token.FileSet) {
	var files []*File
	var astFiles []*ast.File
	var funcUsingCollection = make(map[string]string, 0)

	for _, name := range filePkgPathAndNames {
		data, err := ioutil.ReadFile(name)
		if err != nil {
			// Warn but continue to next package.
			fmt.Printf("1 %s: %s", name, err)
			return nil, nil, fs
		}
		var parsedFile *ast.File
		if strings.HasSuffix(name, ".go") {
			parsedFile, err = parser.ParseFile(fs, name, data, parser.ParseComments)
			if err != nil {
				fmt.Printf("2=========== %s: %s", name, err)
				return nil, nil, fs
			}
			astFiles = append(astFiles, parsedFile)
		}
		files = append(files, &File{
			fset:                fs,
			content:             data,
			name:                name,
			file:                parsedFile,
			funcUsingCollection: funcUsingCollection,
		})
	}
	return files, astFiles, fs
}

// doPackage analyzes the single package constructed from the named files.
// It returns the parsed Package or nil if none of the files have been checked.
func doPackage(files []*File, astFiles []*ast.File, fset *token.FileSet) *Package {

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
	pkg.collFieldTypes = make(map[string]string)
	// Type check the package.
	conf := &types.Config{
		Error: func(e error) {
			fmt.Println("failed to typecheck: ", e)
			os.Exit(1)
		},
		Importer: importer.Default(),
	}
	_, err := conf.Check(pkg.path, fset, astFiles, &info)
	if err != nil {
		fmt.Printf("33 ========== %s\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		file.pkg = pkg
		if file.file != nil {
			file.walkFile(file.name, file.file)
		}
	}
	fmt.Println("results:")
	for k, v := range pkg.collFieldTypes {
		fmt.Printf("k: %s, v: %s\n", k, v)
	}

	return pkg
}
*/
