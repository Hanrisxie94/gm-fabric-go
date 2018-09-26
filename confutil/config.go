package confutil

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
)

// CreateConfigFromBase64 will read in a base64 string from an env var and create a file with its contents decoded
func CreateConfigFromBase64(envVar, path string) error {
	ev := strings.TrimSpace(os.Getenv(envVar))
	if ev == "" {
		log.Println("No value found in environment variable: " + envVar)
		return nil
	}

	data, err := base64.StdEncoding.DecodeString(ev)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0755)
}

// CreateConfigFromTemplate creates a config file by filling a template
// using environment variables.
func CreateConfigFromTemplate(templateText string, path string) error {
	// First we create a FuncMap with which to register the function.
	funcMap := template.FuncMap{
		// The name "getenv" is what the function will be called in the template text.
		"getenv": os.Getenv,
	}

	tmpl, err := template.New("config").Funcs(funcMap).Parse(templateText)
	if err != nil {
		return fmt.Errorf("template.Parse failed: %v: \n%v\n", err, templateText)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("os.Create(%s) failed: %v", path, err)
	}

	err = tmpl.Execute(f, nil)
	if err != nil {
		return fmt.Errorf("tmpl.Execute failed: %v", err)
	}

	return f.Close()
}
