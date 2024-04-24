package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/yargevad/filepathx"
	"gopkg.in/yaml.v3"
)

type Metadata struct {
	ID              string   `yaml:"id"`
	IsCorsResource  *bool    `yaml:"isCorsResource,omitempty"`
	Title           string   `yaml:"title"`
	Subtitle        string   `yaml:"subtitle,omitempty"`
	Permalink       string   `yaml:"permalink,omitempty"`
	Roles           string   `yaml:"roles,omitempty"`
	IsPublic        *bool    `yaml:"isPublic,omitempty"`
	RedirectFrom    []string `yaml:"redirect_from,omitempty"`
	RedirectFromIds []string `yaml:"redirectFromIds,omitempty"`
	AllowedRoles    []string `yaml:"allowedRoles,omitempty"`
	FreeTrialUrl    string   `yaml:"freeTrialUrl,omitempty"`
	HelpKeys        []string `yaml:"helpKeys,omitempty"`
	Next            string   `yaml:"next,omitempty"`
	Previous        string   `yaml:"prev,omitempty"`
	Public          *bool    `yaml:"public,omitempty"`
}

type MarkdownFile struct {
	MetadataPart Metadata
	MarkdownPart string
}

// TODO: make these configurable
var basePath = "/devtools-html/4.0.0/en/"
var filesDir = "/Users/erikla-deque/src/product-docs-site/packages/devtools-html/content/4.0.0/en"

func main() {

	var markdownFile MarkdownFile

	files, err := filepathx.Glob(path.Join(filesDir, "/**/*.md"))

	if err != nil {
		log.Fatal("Could not find Markdown files.")
	}

	for _, file := range files {
		// Split it into YAML and Markdown
		markdownFile = readMarkdownFile(file)

		// Calculate the redirect value

		markdownFile.MetadataPart.RedirectFrom = append(markdownFile.MetadataPart.RedirectFrom, basePath+markdownFile.MetadataPart.ID)
		writeMarkdownFile(file, markdownFile)
	}

}

func readMarkdownFile(filename string) MarkdownFile {

	var markdownFile MarkdownFile

	readFile, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}

	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	currentFileContents := ""

	for fileScanner.Scan() {

		currentLine := fileScanner.Text()

		// Need to test whether the ID is already set because some files contain
		// "---" in code snippets, which would cause the Markdown section to be incorrect.
		if strings.HasPrefix(currentLine, "---") && markdownFile.MetadataPart.ID == "" {
			if currentFileContents != "" {
				yaml.Unmarshal([]byte(currentFileContents), &markdownFile.MetadataPart)
			}
			currentFileContents = ""
		} else {
			currentFileContents += currentLine + "\n"
		}

	}
	markdownFile.MarkdownPart = currentFileContents

	return markdownFile

}

func writeMarkdownFile(filename string, markdownFile MarkdownFile) {

	// Create temporary files
	tempFilename1 := filename + randomString(5)
	tempFilename2 := filename + randomString(5)

	file, err := os.Create(tempFilename1)
	yamlEncoder := yaml.NewEncoder(file)
	yamlEncoder.SetIndent(2)

	defer yamlEncoder.Close()

	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	file.WriteString("---\n")

	yamlEncoder.Encode(markdownFile.MetadataPart)

	file.WriteString("---\n")

	file.WriteString(markdownFile.MarkdownPart)

	yamlEncoder.Close()
	file.Close()

	err = os.Rename(filename, tempFilename2) // Rename original file to temporary file

	if err != nil {
		log.Fatal(err)
	}

	err = os.Rename(tempFilename1, filename) // Rename new file to original file

	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(tempFilename2) // Finally, remove the original file

	if err != nil {
		log.Fatal(err)
	}

}

func randomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}
