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
)

type Config struct {
	Languages []*Language
}

type Language struct {
	Name       string `toml:"name"`
	File       string `toml:"file"`
	Dir        string `toml:"dir"`
	CompileCmd string `toml:"compile"`
	CompileRes template.HTML
	VersionCmd string `toml:"version"`
	VersionRes string
	Webpage    string `toml:"webpage"`
}

func (l *Language) Prepare() {
}

func executeCmd(command string, dir string) (string, error) {
	// TODO: This should return an error if the programm cannot be found but
	// should be fine with the command returning not null
	parts := strings.Split(command, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	f, _ := pty.Start(cmd)
	output, _ := io.ReadAll(f)
	return strings.TrimSpace(string(output)), nil
}

func ansiEscToHTML(style *Formatting, funcName string, params string) string {
	if funcName != "m" {
		fmt.Println("Unknown function:", funcName, params)
		return ""
	}

	// A article on what those numbers mean can be found here
	// https://notes.burke.libbey.me/ansi-escape-codes/
	ansiTable := map[string]func(){
		"":  style.Reset,
		"0": style.Reset,
		"1": func() { style.Bold = true },
		"3": func() { style.Italic = true },
		"4": func() { style.Underline = true },

		"30": func() { style.Color = "" },
		"31": func() { style.Color = "red" },
		"32": func() { style.Color = "green" },
		"33": func() { style.Color = "yellow" },
		"34": func() { style.Color = "blue" },
		"35": func() { style.Color = "purple" },
		"36": func() { style.Color = "cyan" },

		"90": func() { style.Color = "" },
		"91": func() { style.Color = "red" },
		"92": func() { style.Color = "green" },
		"93": func() { style.Color = "yellow" },
		"94": func() { style.Color = "blue" },
		"95": func() { style.Color = "purple" },
		"96": func() { style.Color = "cyan" },

		// TODO: fix this:
		// Hardcode the most common ansi256 colors
		"38;5;9":  func() { style.Color = "red" },
		"38;5;10": func() { style.Color = "green" },
		"38;5;11": func() { style.Color = "yellow" },
		"38;5;12": func() { style.Color = "blue" },
	}

	args := strings.Split(params, ";")
	if len(args) == 0 {
		args = append(args, "")
	}

	current := make([]string, 0)
	for _, arg := range args {
		arg = strings.TrimLeft(arg, "0")
		current = append(current, arg)
		callback, ok := ansiTable[strings.Join(current, ";")]
		if !ok {
			continue
		}

		callback()
		current = nil
	}

	if current != nil {
		fmt.Printf("Unknown ansi code: [%sm\n", strings.Join(current, ";"))
	}

	return style.GenerateHTML()
}

type Formatting struct {
	Color     string
	Bold      bool
	Italic    bool
	Underline bool
}

func (f *Formatting) Reset() {
	f.Color = ""
	f.Bold = false
	f.Italic = false
	f.Underline = false
}

func (f *Formatting) GenerateHTML() string {
	classes := []string{}
	if f.Color != "" {
		classes = append(classes, f.Color)
	}
	if f.Bold {
		classes = append(classes, "bold")
	}
	if f.Italic {
		classes = append(classes, "italic")
	}
	if f.Underline {
		classes = append(classes, "underline")
	}

	return fmt.Sprintf("</span><span class=\"%s\">", strings.Join(classes, " "))
}

func parseAnsiText(ansi string) string {

	style := Formatting{}
	result := ""
	escState := false
	escSeq := ""
	for _, c := range []byte(ansi) {
		if !escState && c != '\x1b' {
			result += string(c)
			continue
		}
		if !escState && c == '\x1b' {
			escState = true
			continue
		}

		if strings.Contains("ABCDEFGHIJKSTsufm", string(c)) {
			if escSeq[0] == '[' {
				result += ansiEscToHTML(&style, string(c), escSeq[1:])
			}
			escState = false
			escSeq = ""
			continue
		}
		escSeq += string(c)
	}

	return result
}

// https://notes.burke.libbey.me/ansi-escape-codes/
func ansiToHTML(ansi string) template.HTML {
	// Escape the output to avoid injections
	ansi = template.HTMLEscapeString(ansi)
	result := "<span>"
	result += parseAnsiText(ansi)
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

		lang.VersionRes, err = executeCmd(lang.VersionCmd, "")
		if err != nil {
			msg := fmt.Sprintf("Unable to execute `%s` %s:", lang.VersionCmd, lang.Name)
			log.Fatal(msg, err.Error())
		}

		compileOutput, err := executeCmd(lang.CompileCmd, lang.Dir)
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
