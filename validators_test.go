package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateNoCircular(t *testing.T) {
	tests := []struct {
		stack   string
		process *Process
	}{
		{
			stack: "",
			process: &Process{
				Name: "one",
				Before: []*Process{
					{
						Name: "foo",
					},
				},
			},
		},
		{
			stack: "foo\nbar",
			process: &Process{
				Name: "foo",
				Before: []*Process{
					{
						Name: "bar",
						Before: []*Process{
							{
								Name: "foo",
							},
						},
					},
				},
			},
		},
	}

	a := assert.New(t)

	for _, test := range tests {
		stack := ValidateNoCircular(test.process, "")
		a.Equal(test.stack, stack)
	}
}

func TestValidateNoDuplicates(t *testing.T) {
	a := assert.New(t)

	tests := []struct {
		hasError  bool
		processes []*Process
	}{
		{
			hasError: false,
			processes: []*Process{
				{
					Name: "foo",
				},
				{
					Name: "Bar",
				},
				{
					Name: "baz",
				},
			},
		},
		{
			hasError: true,
			processes: []*Process{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
				{
					Name: "baz",
				},
				{
					Name: "foo",
				},
			},
		},
	}

	for _, test := range tests {
		err := ValidateNoDuplicates(test.processes)
		if test.hasError {
			a.Error(err)
		} else {
			a.NoError(err)
		}
	}
}

func TestValidateNoDependOnAutoRestart(t *testing.T) {
	a := assert.New(t)

	tests := []struct {
		hasError bool
		process  *Process
	}{
		{
			hasError: false,
			process: &Process{
				Name:        "foo",
				AutoRestart: true,
			},
		},
		{
			hasError: false,
			process: &Process{
				Name:        "bar",
				AutoRestart: false,
				Before:      []*Process{{Name: "test"}},
			},
		},
		{
			hasError: true,
			process: &Process{
				Name:        "baz",
				AutoRestart: true,
				Before:      []*Process{{Name: "test2"}},
			},
		},
	}

	for _, test := range tests {
		err := ValidateNoDependOnAutoRestart(test.process)
		if test.hasError {
			a.Error(err)
		} else {
			a.NoError(err)
		}
	}
}
