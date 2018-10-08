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
	var funcMap template.FuncMap = template.FuncMap{
		// getenv will return the contents of an environment variable
		// it will return an empty string if the environment variable does
		// not exist
		"getenv": os.Getenv,

		// getbool will return false unless the environment variable is exactly
		// equal to 'true'
		"getbool": func(key string) bool { return os.Getenv(key) == "true" },
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
