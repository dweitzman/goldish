// Package goldish helps with golden testing when you want many tests encoded in a
// single file in a human-friendly encoding.
package goldish

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "Update golden files")

type testCase struct {
	Comment string
	Args    map[string]string
}

func (e *testCase) toString(keyOrder []string) (string, error) {
	bw := bytes.NewBuffer(nil)

	if e.Comment != "" {
		bw.WriteString(e.Comment)
		bw.WriteString("\n")
	}
	for _, key := range keyOrder {
		value, ok := e.Args[key]
		if !ok {
			continue
		}
		if value == "" {
			continue
		}
		bw.WriteString(key)
		bw.WriteString(":\n")
		bw.WriteString("  ")
		bw.WriteString(strings.Replace(value, "\n", "\n  ", -1))
		bw.WriteString("\n")
	}

	return bw.String(), nil
}

type cases struct {
	All []testCase
}

func (e *cases) Write(w io.Writer, keyOrder []string) error {
	for i, one := range e.All {
		oneStr, err := one.toString(keyOrder)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, oneStr)
		if err != nil {
			return err
		}

		if i+1 < len(e.All) {
			_, err = w.Write([]byte{'\n'})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// parse scans text in the Goldish expectation format.
// The keys arguments is an ordered list of exactly which
// keys must be present in each expectation.
func parse(r io.Reader) (*cases, error) {
	sc := bufio.NewScanner(r)
	buf := bytes.NewBuffer(nil)

	var result []testCase

	// Read cases
	for sc.Text() != "" || sc.Scan() {
		ex := testCase{Comment: "", Args: make(map[string]string)}

		for strings.HasPrefix(sc.Text(), "#") {
			if buf.Len() > 0 {
				buf.WriteString("\n")
			}
			buf.WriteString(sc.Text())
			sc.Scan()
		}
		ex.Comment = buf.String()
		buf.Reset()

		for strings.HasSuffix(sc.Text(), ":") {
			// It's a key
			keyName := sc.Text()
			keyName = keyName[:len(keyName)-1]
			sc.Scan()
			for strings.HasPrefix(sc.Text(), "  ") {
				if buf.Len() > 0 {
					buf.WriteString("\n")
				}
				buf.WriteString(sc.Text()[2:])
				sc.Scan()
			}
			ex.Args[keyName] = buf.String()
			buf.Reset()
		}

		result = append(result, ex)

		// Skip blank lines
		for sc.Text() == "" && sc.Scan() {
		}
	}

	return &cases{result}, nil
}

// CaseEvaluator is a user-provided function that computes the actual outputs
// based on the inputs for a particular test. CheckGolden() will compare the
// computed outputs with the saved outputs, or simply save them if the
// -update command line flag is set.
type CaseEvaluator func(in map[string]string) (outs map[string]string, err error)

// CheckGoldens executes a golden test.
//
// The default behavior is to verify the correctness of the golden
// output. Setting the -update flag will overwrite the test cases file
// with updated golden output:
//
//   $ go test ./ -update
//
// Test cases and arguments are read from the file testdata/<TestName>_cases.txt
//
// When writing golden output, the keys for each test case will be reordered
// so that the inKeys appear in order followed by the outKeys.
func CheckGoldens(t *testing.T, inKeys []string, outKeys []string, evaluator CaseEvaluator) {
	fname := "testdata/" + t.Name() + "_cases.txt"
	f, err := os.Open(fname)
	if err != nil {
		t.Fatal(err)
	}
	cases, err := parse(f)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases.All {
		savedOuts := make(map[string]string)
		for _, k := range outKeys {
			oldVal, ok := c.Args[k]
			if ok {
				savedOuts[k] = oldVal
			}
			delete(c.Args, k)
		}

		newOuts, err := evaluator(c.Args)
		if err != nil {
			t.Fatal(err)
		}

		for k, v := range newOuts {
			isValid := false
			for _, validKey := range outKeys {
				if validKey == k {
					isValid = true
				}
			}
			if !isValid {
				t.Fatalf("%v is not a valid out key", k)
			}
			c.Args[k] = v
		}

		if !*update {
			// Check for any changes
			if !reflect.DeepEqual(savedOuts, newOuts) {
				t.Errorf("Update golden file with -update. %v vs %v", savedOuts, newOuts)
				continue
			}
		}
	}

	f.Close()

	if !*update {
		return
	}

	allKeys := append([]string{}, inKeys...)
	allKeys = append(allKeys, outKeys...)

	b := bytes.NewBuffer(nil)
	err = cases.Write(b, allKeys)
	if err != nil {
		t.Fatal(err)
	}

	w, err := os.Create(fname)
	if err != nil {
		t.Fatal(err)
	}

	_, err = w.Write(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	defer w.Close()
}
