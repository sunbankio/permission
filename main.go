package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	plugin "github.com/zeromicro/go-zero/tools/goctl/plugin"

	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
)

var (
	handlerdir = flag.String("handlerdir", "", "the directory of the handler files")
	tplFile    = flag.String("tpl", "", "the template file")
	typesPkg   = flag.String("types", "", "the type package containing the context keys")
	tpl        = ""
)

func getHandlerBaseName(route spec.Route) (string, error) {
	handler := route.Handler
	handler = strings.TrimSpace(handler)
	handler = strings.TrimSuffix(handler, "handler")
	handler = strings.TrimSuffix(handler, "Handler")
	return handler, nil
}
func getHandlerName(route spec.Route, folder string) string {
	handler, err := getHandlerBaseName(route)
	if err != nil {
		panic(err)
	}

	handler = handler + "Handler"
	if folder != *handlerdir {
		handler = strings.Title(handler)
	}
	return handler
}

func replacePermission(file, new string) bool {
	re, err := regexp.Compile(`utils\.HasPermission\(userPerms,"([^"]*)"`)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return false
	}
	matches := re.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		return false
	}
	oldPermission := matches[0][1]

	newContent := re.ReplaceAllString(string(data), fmt.Sprintf(`utils.HasPermission(userPerms,"%s"`, new))

	err = ioutil.WriteFile(file, []byte(newContent), 0644)

	if err != nil {
		return false
	}

	fmt.Println("old permission:", oldPermission, "new permission:", new)

	return true
}

// goctl api plugin -plugin goctl-docplugin="-handlerdir /app/adminapi/internal/handler" -api api/zeroapi/adminapi.api
func main() {

	flag.Parse()

	if *handlerdir == "" {
		fmt.Println("-handlerdir is required")
		return
	}

	if *tplFile == "" {
		fmt.Println("-tpl is required")
		return
	}

	if *typesPkg == "" {
		fmt.Println("-types is required")
		return
	}

	plug, err := plugin.NewPlugin()

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	tplFilePath := filepath.Join(plug.Dir, *tplFile)

	content, err := ioutil.ReadFile(tplFilePath)

	if err != nil {
		fmt.Println(err)
		return
	}

	tpl = string(content)

	service := plug.Api.Service.JoinPrefix()

	fmt.Println("code gen =>", plug.Dir)

	for _, group := range service.Groups {
		for _, route := range group.Routes {

			handlerFilePath := filepath.Join(*handlerdir, group.Annotation.Properties["group"], strings.ToLower(getHandlerName(route, ""))+".go")

			// fmt.Println( handlerdir, group.Annotation.Properties," ==>", handlerFilePath,)

			handlerFinalPath := filepath.Join(plug.Dir, handlerFilePath)

			tags := route.AtDoc.Properties

			if v, ok := tags["permission"]; ok {

				// if isExist := replacePermission(handlerFinalPath, v); isExist {
				// 	fmt.Println("skipping code gen, just replacing permission:", v)
				// 	continue
				// }

				// if err := AddImport(handlerFinalPath, *middleWarePkg); err != nil {
				// 	fmt.Printf("error adding middleWarePkg: %v\n", err)
				// }

				if err := AddImport(handlerFinalPath, "github.com/sunbankio/permission/utils"); err != nil {
					fmt.Printf("error adding utils: %v\n", err)
				}

				if err := AddImport(handlerFilePath, *typesPkg); err != nil {
					fmt.Printf("error adding types package: %v\n", err)
				}

				fmt.Println("modifying =>", handlerFinalPath)
				fmt.Println("permission:", v)

				if err := AppendAfterParse(handlerFinalPath, fmt.Sprintf(tpl, v)); err != nil {
					fmt.Printf("error adding code: %v\n", err)
				}
			}

		}
	}

}

// if v, ok := tags["import"]; ok {
// 	fmt.Println("import:", v)
// 	if err := AddImport(handlerFinalPath, v); err != nil {
// 		fmt.Printf("error adding import: %v\n", err)
// 	}
// }

func AddImport(filename string, importPath string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Find the import declaration
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.IMPORT {
			continue
		}

		// Add new import spec
		newImport := &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: `"` + importPath + `"`,
			},
		}
		genDecl.Specs = append(genDecl.Specs, newImport)
		break
	}

	// Write changes back to file
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return printer.Fprint(f, fset, file)
}

// AddCodeAfterParseBlock adds code after the httpx.Parse block
func AddCodeAfterParseBlock(filePath string, newCode string) error {
	fset := token.NewFileSet()

	// Parse the file
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %v", err)
	}

	// Find the function
	ast.Inspect(file, func(n ast.Node) bool {
		// Look for the function declaration
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != "CreateChannelHandler" {
			return true
		}

		fmt.Println("found CreateChannelHandler")
		// Find the if statement with httpx.Parse
		ast.Inspect(funcDecl, func(n ast.Node) bool {

			fmt.Println("funcDecl.Body.List:", len(funcDecl.Body.List))

			ifStmt, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			sel, ok := ifStmt.X.(*ast.Ident)
			if !ok {
				return true
			}

			if sel.Name != "httpx" {
				return true
			}

			if ifStmt.Sel.Name != "Parse" {
				return true
			}

			fmt.Println("found if statement")

			//get the end position of the if statement and insert new code
			endPos := ifStmt.Pos()
			fmt.Println("endPos:", endPos)

			// Create your new statement
			newStmt := &ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.Ident{
						Name: newCode,
					},
					Args: []ast.Expr{},
				},
			}

			insertStmt(funcDecl.Body, endPos, newStmt)

			return false
		})
		return false
	})

	// Write the modified AST back to file
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = format.Node(f, fset, file)
	if err != nil {
		panic(err)
	}

	return nil
}

func insertStmt(block *ast.BlockStmt, pos token.Pos, newStmt ast.Stmt) {
	// Find insert index by comparing positions
	index := 0
	for i, stmt := range block.List {
		if stmt.Pos() >= pos {
			index = i
			break
		}
	}

	fmt.Println("index:", index, len(block.List))

	// Create new slice with extra capacity
	newList := make([]ast.Stmt, len(block.List)+1)

	// Copy statements before insert point
	copy(newList[:index], block.List[:index])

	// Insert new statement
	newList[index] = newStmt

	// Copy remaining statements
	copy(newList[index+1:], block.List[index:])

	// Update block's statement list
	block.List = newList
}

func AppendAfterParse(filePath string, appendStr string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Define regex pattern to match the Parse block
	pattern := regexp.MustCompile(`if err := httpx\.Parse\(r, &req\); err != nil \{[^}]+\}`)

	// Find the match in the content
	loc := pattern.FindIndex(content)
	if loc == nil {
		return fmt.Errorf("pattern not found in file")
	}

	// Split content into before and after the match
	before := string(content[:loc[1]])
	after := string(content[loc[1]:])

	// Add a newline if appendStr doesn't start with one
	if !strings.HasPrefix(appendStr, "\n") {
		appendStr = "\n" + appendStr
	}

	// Combine the parts with the new string
	newContent := before + appendStr + after

	// Write back to file
	return os.WriteFile(filePath, []byte(newContent), 0644)
}
