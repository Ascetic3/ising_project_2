package csvio

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDemoInputMatchesDocumentedSchema(t *testing.T) {
	expectedHeader := []string{
		"L", "J1", "J2", "J3", "J4", "J5", "J6",
		"copies", "h", "T", "aSteps", "mSteps", "save",
	}
	root := filepath.Join("..", "..")
	demoPath := filepath.Join(root, "configs", "demo-input.csv")

	file, err := os.Open(demoPath)
	if err != nil {
		t.Fatalf("open demo input: %v", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ';'
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("read demo input: %v", err)
	}
	if len(records) < 2 {
		t.Fatalf("demo input has %d rows, want header and at least one point", len(records))
	}
	if !reflect.DeepEqual(records[0], expectedHeader) {
		t.Fatalf("demo header = %v, want %v", records[0], expectedHeader)
	}

	for rowIndex, record := range records {
		params, skip, err := ParseRecord(record, rowIndex)
		if err != nil {
			t.Fatalf("parse demo row %d: %v", rowIndex+1, err)
		}
		if rowIndex == 0 {
			if !skip {
				t.Fatal("parser did not recognize the demo header")
			}
			continue
		}
		if skip {
			t.Fatalf("parser unexpectedly skipped demo row %d", rowIndex+1)
		}
		if len(record) != len(expectedHeader) {
			t.Fatalf("demo row %d has %d fields, want %d", rowIndex+1, len(record), len(expectedHeader))
		}
		if rowIndex == 1 {
			want := Params{
				L: 12, J1: 1, J2: 1, J3: 1, J4: 1, J5: 1, J6: 1,
				Copies: 2, H: 0, T: 0.5, ASteps: 100, MSteps: 200, Save: true,
			}
			if params != want {
				t.Fatalf("first demo point = %+v, want %+v", params, want)
			}
		}
	}

	documentedHeader := strings.Join(expectedHeader, ";")
	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if !strings.Contains(string(readme), documentedHeader) {
		t.Fatalf("README does not contain the supported input header %q", documentedHeader)
	}
	if strings.Contains(string(readme), "J6;K;copies") {
		t.Fatal("README still documents unsupported K input column")
	}
}
