package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/creack/pty"
	"github.com/jwalton/go-ansiparser"
)

type Config struct {
	Languages []*Language
}

type Language struct {
	Name       string `toml:"name"`
	File       string `toml:"file"`
	CompileCmd string `toml:"compile"`
	CompileRes template.HTML
	VersionCmd string `toml:"version"`
	VersionRes string
	Webpage    string `toml:"webpage"`
}

func (l *Language) Prepare() {
}

func executeCmd(command string) (string, error) {
	// TODO: This should return an error if the programm cannot be found but
	// should be fine with the command returning not null

	parts := strings.Split(command, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	f, _ := pty.Start(cmd)
	output, _ := io.ReadAll(f)
	return strings.TrimSpace(string(output)), nil
}

func ansiToHTML(ansi string) template.HTML {
	// Escape the output to avoid injections
	ansi = template.HTMLEscapeString(ansi)

	// Parse the ansi
	tokenizer := ansiparser.NewStringTokenizer(ansi)

	ansiTable := map[string]string{
		"":   "",
		"30": "",
		"31": "red",
		"32": "green",
		"33": "yellow",
		"34": "blue",
		"35": "purple",

		// TODO: fix this:
		// Hardcode the most common ansi256 colors
		"38;5;9":  "red",
		"38;5;10": "green",
		"38;5;11": "yellow",
		"38;5;12": "blue",
	}

	// TODO: this is ugly .. fix it
	result := "<span>"
	for tokenizer.Next() {
		token := tokenizer.Token()
		if token.Type == ansiparser.String {
			result += token.Content
			continue
		}

		// token is a command
		class, ok := ansiTable[token.FG]
		if !ok {
			fmt.Printf("%+v\n", token.FG)
			class = ""
		}

		result += fmt.Sprintf("</span><span class=\"%s\">", class)
	}
	result += "</span>"

	return template.HTML(result)
}

func main() {
	var config Config
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, lang := range config.Languages {
		lang.CompileCmd = strings.ReplaceAll(lang.CompileCmd, "$FILE", lang.File)
		log.Printf("Running %s: %s", lang.Name, lang.CompileCmd)

		lang.VersionRes, err = executeCmd(lang.VersionCmd)
		if err != nil {
			msg := fmt.Sprintf("Unable to execute `%s` %s:", lang.VersionCmd, lang.Name)
			log.Fatal(msg, err.Error())
		}

		compileOutput, err := executeCmd(lang.CompileCmd)
		lang.CompileRes = ansiToHTML(compileOutput)
		if err != nil {
			msg := fmt.Sprintf("Unable to execute `%s` %s:", lang.CompileCmd, lang.Name)
			log.Fatal(msg, err.Error())
		}
	}

	temp := template.Must(template.ParseFiles("template.html"))
	file, err := os.Create("index.html")
	if err != nil {
		log.Fatal("Unable to create index.html", err.Error())
	}

	err = temp.Execute(file, config.Languages)
	if err != nil {
		log.Fatal("Cannot render template", err.Error())
	}
}
