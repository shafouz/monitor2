package utils_test

import (
	"bytes"
	"html/template"
	"monitor2/utils"
	"os"
	"testing"

	"github.com/rs/zerolog/log"
)

func TestTemplates(t *testing.T) {
  templates := template.Must(template.ParseGlob("../static/templates/*.html"))
  log.Printf("DEBUGPRINT[3]: utils_test.go:15: templates.DefinedTemplates=%+v\n", templates.DefinedTemplates())
}

func TestSplitTerminatorRemovesAll(t *testing.T) {
	d := []byte("\n\n\n\n")
	a := utils.SplitTerminator(d, "\n")
	if len(a) != 0 {
		t.Fatal()
	}
}

func TestDiffEmpty(t *testing.T) {
	d1 := []byte("")
	d2 := []byte("asdasd\nasdasd")
	d := utils.Diff("old", d1, "new", d2)

	if len(d) == 0 {
		t.Fatal()
	}
}

func TestRunScript(t *testing.T) {
	os.WriteFile("/tmp/tmp123.py", []byte(`import sys
req = sys.stdin.read()
print(req)`), 0755)
	b, err := utils.RunPyScript("/tmp/tmp123.py", []byte("hello"), []string{})
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(b, []byte("hello\n")) != 0 {
		t.Fatal(err)
	}
}

func TestSortBytes(t *testing.T) {
	arr := [][]byte{
		[]byte("zAbcedf"),
		[]byte("0"),
		[]byte("zzAbcedf"),
		[]byte("Abcedf"),
		[]byte("abcedf"),
	}

	utils.SortBytes(arr)
	if bytes.Compare(arr[0], []byte("0")) != 0 {
		t.Fatal()
	}
}

func TestCompactBytes(t *testing.T) {
	arr := [][]byte{
		[]byte("zAbcedf"),
		[]byte("0"),
		[]byte("zzAbcedf"),
		[]byte("Abcedf"),
		[]byte("abcedf"),
		[]byte("abcedf"),
	}

	prev_len := len(arr)
	arr = utils.CompactBytes(arr)
	if len(arr) == prev_len {
		t.Fatal()
	}
}
