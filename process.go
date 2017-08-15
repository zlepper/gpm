package main

import (
	"context"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Describes a single process
type Process struct {
	// The name of the process
	Name string
	// The command to run
	Command string
	// The args to pass along to the command
	Args []string
	// If the process should be automatically restarted if it stops
	AutoRestart bool
	// Proccesses that should be run after this process
	Before []*Process
	// The name of the process, this process should run after
	after string
}

func (p *Process) restart(ctx context.Context) (err error) {
	log.Printf("Process '%s' has died an untimely death, and will be restarted.\n", p.Name)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = p.Run(ctx)
	}()
	wg.Wait()
	return err
}

// Runs the actual process
func (p *Process) Run(ctx context.Context) error {
	log.Printf("Starting process '%s'\n", p.Name)

	command, err := exec.LookPath(p.Command)
	if err != nil {
		return err
	}

	cmd := exec.Command(command, p.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	done := make(chan error, 1)
	// We have to reimplement all this, as os.Process does a SIGKILL, and we want to give the process a
	// change to exit gracefully

	err = cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		cmd.Process.Signal(os.Interrupt)
		select {
		// Give the process 7 seconds to kill itself, before we kill it
		case <-time.After(time.Second * 7):
			log.Printf("Process '%s' is stuborn. Sending SIGKILL", p.Name)
			return cmd.Process.Kill()
		case <-done:
			log.Printf("Process '%s' has finished gracefully\n", p.Name)

			// We are existing gracefully, and doesn't have any error
			return nil
		}
	case err = <-done:
		// The process has finished. Check if we should start it again
		if p.AutoRestart {
			return p.restart(ctx)
		} else {
			log.Printf("Process '%s' has finished\n", p.Name)
		}
	}

	if err != nil {
		return err
	}

	if len(p.Before) > 0 {
		var g errgroup.Group

		for _, process := range p.Before {
			process := process
			g.Go(func() error {
				return process.Run(ctx)
			})
		}

		return g.Wait()
	}

	return nil
}
