package utils

import (
	"bytes"
	"errors"

	// "fmt"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/rs/zerolog/log"
)

func LogErr(err error) {
  log.Err(err).Caller().Caller().Msg("")
}

// removes the suffix term after splitting
func SplitTerminator(arr []byte, term string) [][]byte {
  trimmed := bytes.TrimRight(arr, term)
  if len(trimmed) == 0 { return nil }
  return bytes.Split(trimmed, []byte(term))
}

func CompactBytes(arr [][]byte) [][]byte {
  return slices.CompactFunc(arr, func(a []byte, b []byte) bool {
    if bytes.Compare(a, b) == 0 {
      return true
    }
    return false
  })
}

func SortBytes(arr [][]byte){
  slices.SortFunc(arr, func(a []byte, b []byte) int { 
    return bytes.Compare(a, b)
  })
}

func RunPyScript(abs_path string, _stdin []byte, extra_args []string) ([]byte, error) {
	if !filepath.IsAbs(abs_path) {
		return nil, errors.New("Path not absolute.")
	}

	var outb, errb bytes.Buffer

  args := []string{
    abs_path,
    // fmt.Sprint(len(_stdin)),
  }
  args = append(args, extra_args...)

  log.Printf("python3 %+v\n", args)
	cmd := exec.Command("python3", args...)
	cmd.Stderr = &errb
	cmd.Stdout = &outb
  cmd.Stdin = bytes.NewReader(_stdin)

	cmd.Run()
  if errb.Len() != 0 {
    return nil, errors.New(errb.String())
  }

  return outb.Bytes(), nil
}
