package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTokenize(t *testing.T) {
	s := `gfs --config=/foo/var --serve=\"/bar\ baz/foo\" -d /var\ foo "var foo" C:\this\is\test`
	expected := []string{`gfs`, `--config=/foo/var`, `--serve="/bar baz/foo"`, "-d", "/var foo", "var foo", `C:\this\is\test`}

	actual := Tokenize(s)

	assert.Equal(t, expected, actual)
}

func TestProcessManager_buildOneProcess(t *testing.T) {
	pj := processJson{
		Name:        "test",
		AutoRestart: true,
		Command:     `foo bar "baz bin" bin\ boo`,
		After:       "foo",
	}

	pm := ProcessManager{}
	process, err := pm.buildOneProcess(&pj)

	a := assert.New(t)

	if a.NoError(err) {

		a.Equal("test", process.Name)
		a.Equal(true, process.AutoRestart)
		a.Equal("foo", process.Command)
		a.Equal("foo", process.after)
		a.Equal([]string{"bar", "baz bin", "bin boo"}, process.Args)
	}
}
