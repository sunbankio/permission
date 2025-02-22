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

				if err := AddImport(handlerFinalPath, "utils", "github.com/sunbankio/permission/utils"); err != nil {
					fmt.Printf("error adding utils: %v\n", err)
				}

				if err := AddImport(handlerFinalPath, "contextkey", *typesPkg); err != nil {
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

func AddImport(filename string, alias, importPath string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	if alias != "" {
		alias += " "
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
				Value: alias + `"` + importPath + `"`,
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
