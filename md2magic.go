// main.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

func absDirPath(path string) (string, error) {
	absInputDir, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("Error getting working directory: %v\n", err)
	}
	fileInfo, err := os.Stat(absInputDir)
	if err != nil {
		return "", fmt.Errorf("Error stat-ing %s: %v\n", absInputDir, err)
	}
	if !fileInfo.IsDir() {
		return "", fmt.Errorf("%s is not a directory\n", absInputDir)
	}
	return absInputDir, nil
}

func main() {
	fmt.Println("Markdown to HTML Converter")
	// Define command-line flags
	inputDir := flag.String("input", "./md", "Path to the directory containing input markdown files")
	outputDir := flag.String("output", "./out", "Path to the output directory of HTML files")
	templateFileName := flag.String("template", "template.html", "filename of html template in input directory")
	defaultTitle := flag.String("title", "Document", "Default title of the HTML document")
	flag.Parse()

	// Check if input and output directories exists
	absInputDir, err := absDirPath(*inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting absolute path: %v", err)
		os.Exit(1)
	}
	absOutputDir, err := absDirPath(*outputDir)
	if err != nil {
		// attempt to create output
		err := os.MkdirAll(*outputDir, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v", err)
			os.Exit(1)
		}
		absOutputDir, err = absDirPath(*outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get full filepath after creating file: %s: %v", *outputDir, err)
			os.Exit(1)
		}
	}
	fmt.Printf(" in-dir: %s\n", absInputDir)
	fmt.Printf("out-dir: %s\n", absOutputDir)

	htmlTemplateBytes, err := os.ReadFile(path.Join(absInputDir, *templateFileName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read html template file: %s: %v", templateFileName, err)
		os.Exit(1)
	}

	files, err := ioutil.ReadDir(absInputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list files in %s: %v\n", absInputDir, err)
	}

	extensions := parser.CommonExtensions
	re := regexp.MustCompile(`(?m)^#+\s*([a-zA-Z][a-zA-Z0-9\s]+).*`)
	for _, file := range files {
		filePath := filepath.Join(absInputDir, file.Name())
		// Copy directories from input to output directory
		if file.IsDir() {
			fs := os.DirFS(filePath)
			err := os.CopyFS(absOutputDir, fs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to copy %s from input to output directory: %v", filePath, err)
			}
			continue
		}
		// Copy files from input to output directory
		if filepath.Ext(file.Name()) != ".md" {
			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to copy %s from input to output directory: %v", filePath, err)
				continue
			}
			newFilePath := filepath.Join(absOutputDir, file.Name())
			os.Remove(newFilePath) // We get error if it does not exist.
			err = os.WriteFile(newFilePath, content, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to copy %s from input to output directory: %v", filePath, err)
			}
			continue
		}
		// transpile md -> html
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to read file %s: %v\n", filePath, err)
			continue
		}
		//htmlContent := htmlTemplate
		p := parser.NewWithExtensions(extensions)
		htmlFromMD := markdown.ToHTML(content, p, nil)

		title := *defaultTitle
		match := re.FindSubmatch(content)
		if len(match) > 1 {
			title = string(match[1])
		}
		htmlContent := bytes.Replace(htmlTemplateBytes, []byte("%(CONTENTS)"), htmlFromMD, 1)
		htmlContent = bytes.Replace(htmlContent, []byte("%(TITLE)"), []byte(title), 1)
		htmlFileName := filepath.Join(absOutputDir, strings.Replace(file.Name(), ".md", ".html", 1))
		os.Remove(htmlFileName)
		os.WriteFile(htmlFileName, htmlContent, 0664)
	}
}
