package main

import (
	"bufio"
	"encoding/json"
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
	handlerdir  = flag.String("handlerdir", "", "the directory of the handler files")
	tplFile     = flag.String("tpl", "", "the template file")
	typesPkg    = flag.String("types", "", "the type package containing the context keys")
	tpl         = ""
	dumpDir     = flag.String("dump", "", "the directory of the dump files")
	imports     = flag.String("imports", "", "additional imports. comma separated")
	versionFlag = flag.Bool("version", false, "print version information")
)

const (
	VERSION     = "v1.0.5"
	tplStartTag = "//permission:start"
	tplEndTag   = "//permission:end"
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

	if *versionFlag {
		fmt.Println("Version:", VERSION)
		os.Exit(0)
	}

	dumpOnly := false

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

	if dumpDir != nil {
		if *dumpDir != "" {

			dumpOnly = true
		}
	}

	additionalImports := []string{}

	if *imports != "" {
		additionalImports = strings.Split(*imports, ",")
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

	allPermission := make([]string, 0)

	for _, group := range service.Groups {
		for _, route := range group.Routes {
			tags := route.AtDoc.Properties

			if v, ok := tags["permission"]; ok {
				handlerFilePath := filepath.Join(*handlerdir, group.Annotation.Properties["group"], strings.ToLower(getHandlerName(route, ""))+".go")
				handlerFinalPath := filepath.Join(plug.Dir, handlerFilePath)

				if !dumpOnly {
					if err := AddImport(handlerFinalPath, "utils", "github.com/sunbankio/permission/utils"); err != nil {
						fmt.Printf("error adding utils: %v\n", err)
					}

					if err := AddImport(handlerFinalPath, "contextkey", *typesPkg); err != nil {
						fmt.Printf("error adding types package: %v\n", err)
					}

					for _, imp := range additionalImports {
						if err := AddImport(handlerFinalPath, "", imp); err != nil {
							fmt.Printf("error adding import: %v\n", err)
						}
					}

					fmt.Println("modifying =>", handlerFinalPath)
					fmt.Println("permission:", v)

					content := tplStartTag + "\n" + fmt.Sprintf(tpl, v) + "\n" + tplEndTag

					if err := removeContentBetween(handlerFinalPath, tplStartTag, tplEndTag); err != nil {
						fmt.Printf("error adding code: %v\n", err)
					}

					if err := AppendAfterParse(handlerFinalPath, content); err != nil {
						fmt.Printf("error adding code: %v\n", err)
					}
				}

				allPermission = append(allPermission, v)
			}
		}
	}

	if dumpDir != nil {
		if *dumpDir != "" {

			jsonPermissionBytes, err := json.Marshal(allPermission)

			if err != nil {
				fmt.Println(err)
				return
			}

			filename := filepath.Join(plug.Dir, *dumpDir, "permission.json")

			fmt.Println("dump =>", filename)

			if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
				fmt.Println(err)
				return
			}

			f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer f.Close()

			_, err = f.Write(jsonPermissionBytes)
			if err != nil {
				fmt.Println(err)
				return
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

	// Check if the import already exists
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		for _, spec := range genDecl.Specs {

			importSpec, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}

			if strings.Contains(importSpec.Path.Value, `"`+importPath+`"`) {
				// Import already exists, do nothing
				fmt.Println("import already exists", importSpec.Path.Value)
				return nil
			}
		}

		// Add new import spec if it doesn't exist
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

func removeContentBetween(filepath, startDelim, endDelim string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	isDeleting := false

	startDelimRegex := regexp.MustCompile(startDelim)
	endDelimRegex := regexp.MustCompile(endDelim)

	for scanner.Scan() {
		line := scanner.Text()

		if startDelimRegex.MatchString(line) {
			isDeleting = true
		}

		if !isDeleting {
			lines = append(lines, line)
		}

		if endDelimRegex.MatchString(line) {
			isDeleting = false
		}

	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Write modified content back to the file
	outFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, line := range lines {
		_, err := fmt.Fprintln(writer, line)
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
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
