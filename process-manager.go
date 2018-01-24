package main

import (
	"context"
	"encoding/json"
	"errors"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
)

type processJson struct {
	Name             string `json:"name"`
	Command          string `json:"command"`
	AutoRestart      bool   `json:"autoRestart"`
	After            string `json:"after"`
	WorkingDirectory string `json:"workDir"`
}

// The main process manager, managing processes
type ProcessManager struct {
	processes []*Process
}

func (p *ProcessManager) ParseConfigFile(pathToConfigFile string) error {
	file, err := os.Open(pathToConfigFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var processes []*processJson
	err = json.NewDecoder(file).Decode(&processes)
	if err != nil {
		return err
	}

	log.Println(processes)
	return p.BuildProcesses(processes)
}

// Tokenises the given string as command line arguments
func Tokenize(s string) (tokens []string) {
	// The token that we are currently building
	token := ""

	// True if we are within a qoute
	hasQuote := false
	// True if the next char should be treated as a literal
	escapeNext := false
	for index, char := range s {
		if escapeNext {
			token += string(char)
			escapeNext = false
			continue
		}

		if char == '"' {
			hasQuote = !hasQuote
			continue
		}

		if hasQuote {
			token += string(char)
			continue
		}

		if char == '\\' {
			if index+1 < len(s) {
				nextChar := s[index+1]
				if nextChar == ' ' || nextChar == '"' {
					escapeNext = true
					continue
				}
			}
		}

		if char == ' ' {
			if token != "" {
				tokens = append(tokens, token)
				token = ""
			}
			continue
		}

		token += string(char)

	}

	// Remember to save the last token
	if token != " " {
		tokens = append(tokens, token)
	}

	return
}

func (p *ProcessManager) buildOneProcess(pj *processJson) (*Process, error) {
	process := &Process{
		Name:        pj.Name,
		AutoRestart: pj.AutoRestart,
	}

	tokens := Tokenize(pj.Command)
	process.Command = tokens[0]
	process.Args = tokens[1:]
	process.after = pj.After
	if pj.WorkingDirectory == "" {
		process.WorkingDirectory = "."
	} else {
		process.WorkingDirectory = pj.WorkingDirectory
	}

	return process, nil
}

func (p *ProcessManager) BuildProcesses(processJsons []*processJson) error {
	var processes []*Process
	for _, pj := range processJsons {
		process, err := p.buildOneProcess(pj)
		if err != nil {
			return err
		}
		processes = append(processes, process)
	}

	return p.BuildProcessTree(processes)
}

func (p *ProcessManager) runValidations(processes []*Process) (err error) {
	for _, process := range processes {
		err = ValidateNoDependOnAutoRestart(process)
		if err != nil {
			return
		}
		stack := ValidateNoCircular(process, "")
		if stack != "" {
			return errors.New("Found circular dependency between processes: \n" + stack)
		}
	}
	return nil
}

// Builds a process tree to run, as processes can have inter-dependencies
func (p *ProcessManager) BuildProcessTree(processes []*Process) error {

	err := ValidateNoDuplicates(processes)
	if err != nil {
		return err
	}

	for _, process := range processes {
		if process.after != "" {
			for _, other := range processes {
				if other.Name == process.after {
					other.Before = append(other.Before, process)
				}
			}
		}
	}

	err = p.runValidations(processes)
	if err != nil {
		return err
	}

	// Build the actual tree structure
	var ps []*Process
	for _, process := range processes {
		if process.after == "" {
			ps = append(ps, process)
		}
	}

	// Save the tree
	p.processes = ps

	return nil
}

func (p *ProcessManager) StartProcesses(ctx context.Context) error {

	g, ctx := errgroup.WithContext(ctx)

	for _, process := range p.processes {
		process := process
		g.Go(func() error {
			return process.Run(ctx)
		})
	}

	return g.Wait()
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{}
}
