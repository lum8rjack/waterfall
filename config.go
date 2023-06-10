package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Directory string   `yaml:"directory"`
	Project   string   `yaml:"project"`
	Actions   []Action `yaml:"actions"`
}

type Action struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Command     string   `yaml:"command"`
	Enabled     bool     `yaml:"enabled"`
	Stdin       bool     `yaml:"stdin"`
	Subactions  []Action `yaml:"subactions"`
}

type CustomVars struct {
	Project   string
	Directory string
}

// Create a new config based on the provided YAML file
func NewConfig(file string) (Config, error) {
	var config Config

	// Read in the file
	data, err := os.ReadFile(file)
	if err != nil {
		return config, err
	}

	// Unmarshal the YAML data into the struct
	err = yaml.Unmarshal(data, &config)

	// Check if directory exists
	if config.Directory != "" {
		if _, err := os.Stat(config.Directory); os.IsNotExist(err) {
			return config, errors.New("directory does not exist")
		}
	}

	return config, err
}

// Start the flow
func (c *Config) Start() error {
	// Create a new context
	cv := CustomVars{
		Directory: c.Directory,
		Project:   c.Project,
	}

	ctx := context.WithValue(context.Background(), "customVars", cv)

	// Create a waitgroup based on the number of actions
	var wg sync.WaitGroup
	wg.Add(len(c.Actions))

	// Loop over each action and run them
	if len(c.Actions) > 0 {
		for _, a := range c.Actions {
			// Run in a go routine
			go func(a Action) {
				e := a.Run(ctx, []byte{})
				if e != nil {
					log.Printf("Error running %s - %s\n", a.Name, e)
				}
				wg.Done()
			}(a)
		}
	}

	wg.Wait()
	return nil
}

// Run the action
func (a *Action) Run(ctx context.Context, data []byte) error {
	// Output buffer and error
	var out bytes.Buffer
	var err error

	// Check if enabled
	if a.Enabled {
		// Starting
		log.Printf("Action %s - running\n", a.Name)

		vars, _ := ctx.Value("customVars").(CustomVars)

		var tpl bytes.Buffer
		templ := template.Must(template.New("config").Parse(a.Command))
		err = templ.Execute(&tpl, map[string]interface{}{
			"project":   vars.Project,
			"directory": vars.Directory,
		})
		if err != nil {
			return err
		}

		// Setup the command to run
		args := strings.Split(tpl.String(), " ")
		cmd := exec.Command(args[0])
		if len(args) > 1 {
			cmd = exec.Command(args[0], args[1:]...)
		}

		// Setup input/output
		cmd.Stdout = &out
		if a.Stdin {
			cmd.Stdin = bytes.NewBuffer(data)
		}
		err = cmd.Run()
		if err != nil {
			return err
		}

		log.Printf("Action %s - completed action\n", a.Name)
	}

	// Create a waitgroup based on the number of actions
	var wg sync.WaitGroup
	wg.Add(len(a.Subactions))

	// Loop over subactions to execute
	if len(a.Subactions) > 0 {
		for _, sa := range a.Subactions {
			// Run in a go routine
			go func(sa Action) {
				sa.Run(ctx, out.Bytes())
				wg.Done()
			}(sa)
		}
	}

	wg.Wait()

	if len(a.Subactions) > 0 {
		log.Printf("Action %s - completed subactions\n", a.Name)
	}

	return err
}
