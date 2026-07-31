package main

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const testInput = `L;J1;J2;J3;J4;J5;J6;copies;h;T;aSteps;mSteps;save
8;0.71;-1.13;1.37;-0.89;0.53;-1.61;2;0.27;1.4;20;30;1
8;0.71;-1.13;1.37;-0.89;0.53;-1.61;2;0.27;2.2;20;30;1
`

func writeTestInput(t *testing.T, directory, content string) string {
	t.Helper()
	path := filepath.Join(directory, "input.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write test input: %v", err)
	}
	return path
}

func runTestSimulation(t *testing.T, inputPath, outputDir string, seed int64) {
	t.Helper()
	if err := run(runOptions{
		inputPath:  inputPath,
		outputDir:  outputDir,
		seed:       seed,
		saveImages: true,
	}); err != nil {
		t.Fatalf("run test simulation: %v", err)
	}
}

func physicalHashes(t *testing.T, outputDir string) map[string]string {
	t.Helper()
	paths := []string{
		filepath.Join(outputDir, "output.csv"),
		filepath.Join(outputDir, "result.csv"),
	}
	images, err := filepath.Glob(filepath.Join(outputDir, "images", "*.png"))
	if err != nil {
		t.Fatalf("find PNG files: %v", err)
	}
	if len(images) == 0 {
		t.Fatal("simulation produced no PNG files")
	}
	paths = append(paths, images...)
	sort.Strings(paths)

	hashes := make(map[string]string, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read artifact %s: %v", path, err)
		}
		relative, err := filepath.Rel(outputDir, path)
		if err != nil {
			t.Fatalf("make artifact path relative: %v", err)
		}
		hashes[filepath.ToSlash(relative)] = fmt.Sprintf("%x", sha256.Sum256(content))
	}
	return hashes
}

func TestRunSameSeedProducesIdenticalCSVAndPNG(t *testing.T) {
	root := t.TempDir()
	inputPath := writeTestInput(t, root, testInput)
	firstOutput := filepath.Join(root, "first")
	secondOutput := filepath.Join(root, "second")

	runTestSimulation(t, inputPath, firstOutput, 20260731)
	runTestSimulation(t, inputPath, secondOutput, 20260731)

	firstHashes := physicalHashes(t, firstOutput)
	secondHashes := physicalHashes(t, secondOutput)
	if !reflect.DeepEqual(firstHashes, secondHashes) {
		t.Fatalf("same seed produced different artifacts:\nfirst:  %v\nsecond: %v", firstHashes, secondHashes)
	}
}

func TestRunDifferentSeedChangesPhysicalResult(t *testing.T) {
	root := t.TempDir()
	inputPath := writeTestInput(t, root, testInput)
	firstOutput := filepath.Join(root, "first")
	secondOutput := filepath.Join(root, "second")

	runTestSimulation(t, inputPath, firstOutput, 20260731)
	runTestSimulation(t, inputPath, secondOutput, 20260732)

	firstHashes := physicalHashes(t, firstOutput)
	secondHashes := physicalHashes(t, secondOutput)
	for path, firstHash := range firstHashes {
		if secondHashes[path] != firstHash {
			return
		}
	}
	t.Fatal("different seeds produced identical physical CSV and PNG artifacts")
}

func TestGeneratedCSVSchema(t *testing.T) {
	root := t.TempDir()
	inputPath := writeTestInput(t, root, testInput)
	outputDir := filepath.Join(root, "output")
	runTestSimulation(t, inputPath, outputDir, 20260731)

	readCSV := func(name string) [][]string {
		t.Helper()
		file, err := os.Open(filepath.Join(outputDir, name))
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		defer file.Close()
		reader := csv.NewReader(file)
		reader.Comma = ';'
		records, err := reader.ReadAll()
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return records
	}

	outputRecords := readCSV("output.csv")
	if len(outputRecords) != 2 {
		t.Fatalf("output.csv rows = %d, want 2 data rows without header", len(outputRecords))
	}
	for rowIndex, record := range outputRecords {
		if len(record) != 19 {
			t.Fatalf("output.csv row %d fields = %d, want 19", rowIndex+1, len(record))
		}
	}

	resultRecords := readCSV("result.csv")
	if len(resultRecords) != 2 {
		t.Fatalf("result.csv rows = %d, want 2 data rows without header", len(resultRecords))
	}
	for rowIndex, record := range resultRecords {
		if len(record) != 7 {
			t.Fatalf("result.csv row %d fields = %d, want 7", rowIndex+1, len(record))
		}
	}

	diagnosticsRecords := readCSV("diagnostics.csv")
	wantDiagnosticsHeader := []string{"point", "T", "point_seed"}
	if len(diagnosticsRecords) != 3 {
		t.Fatalf("diagnostics.csv rows = %d, want header and 2 data rows", len(diagnosticsRecords))
	}
	if !reflect.DeepEqual(diagnosticsRecords[0], wantDiagnosticsHeader) {
		t.Fatalf("diagnostics header = %v, want %v", diagnosticsRecords[0], wantDiagnosticsHeader)
	}
}

func TestRunMetadataSanitizesExternalAbsolutePaths(t *testing.T) {
	root := t.TempDir()
	privateDir := filepath.Join(root, "private-user-Alice", "local-worktree")
	if err := os.MkdirAll(privateDir, 0o755); err != nil {
		t.Fatalf("create private test directory: %v", err)
	}
	inputPath := writeTestInput(t, privateDir, testInput)
	outputDir := filepath.Join(privateDir, "public-result")

	if err := run(runOptions{inputPath: inputPath, outputDir: outputDir, seed: 20260731}); err != nil {
		t.Fatalf("run with absolute paths: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(outputDir, "run_metadata.json"))
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	var metadata runMetadata
	if err := json.Unmarshal(content, &metadata); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}

	if metadata.InputFile != "input.csv" || metadata.CLI.InputPath != "input.csv" {
		t.Fatalf("sanitized input paths = %q and %q, want input.csv", metadata.InputFile, metadata.CLI.InputPath)
	}
	if metadata.CLI.OutputDir != "public-result" {
		t.Fatalf("sanitized output directory = %q, want public-result", metadata.CLI.OutputDir)
	}
	for name, value := range map[string]string{
		"input_file": metadata.InputFile,
		"cli.input":  metadata.CLI.InputPath,
		"cli.output": metadata.CLI.OutputDir,
	} {
		if filepath.IsAbs(value) || filepath.VolumeName(value) != "" {
			t.Fatalf("%s contains absolute path %q", name, value)
		}
	}
	metadataText := string(content)
	for _, privateValue := range []string{root, filepath.Base(root), "private-user-Alice", "local-worktree"} {
		if strings.Contains(metadataText, privateValue) {
			t.Fatalf("metadata contains private path component %q", privateValue)
		}
	}
}

func TestRunMetadataPreservesSafeRelativePaths(t *testing.T) {
	root := t.TempDir()
	originalWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("change to temporary working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWorkingDir); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	if err := os.MkdirAll("configs", 0o755); err != nil {
		t.Fatalf("create configs directory: %v", err)
	}
	inputPath := writeTestInput(t, "configs", testInput)
	outputDir := filepath.Join("demo-output", "relative-run")
	if err := run(runOptions{inputPath: inputPath, outputDir: outputDir, seed: 20260731}); err != nil {
		t.Fatalf("run with relative paths: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "run_metadata.json"))
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	var metadata runMetadata
	if err := json.Unmarshal(content, &metadata); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if metadata.InputFile != "configs/input.csv" || metadata.CLI.InputPath != "configs/input.csv" {
		t.Fatalf("relative input paths = %q and %q, want configs/input.csv", metadata.InputFile, metadata.CLI.InputPath)
	}
	if metadata.CLI.OutputDir != "demo-output/relative-run" {
		t.Fatalf("relative output directory = %q, want demo-output/relative-run", metadata.CLI.OutputDir)
	}
	if strings.Contains(string(content), root) || strings.Contains(string(content), filepath.Base(root)) {
		t.Fatalf("metadata contains temporary working directory %q", root)
	}
}

func TestRunRejectsMissingInputBeforeCreatingOutput(t *testing.T) {
	root := t.TempDir()
	outputDir := filepath.Join(root, "output")
	err := run(runOptions{
		inputPath: filepath.Join(root, "missing.csv"),
		outputDir: outputDir,
		seed:      1,
	})
	if err == nil || !strings.Contains(err.Error(), "cannot open input CSV") {
		t.Fatalf("missing input error = %v", err)
	}
	if _, statErr := os.Stat(outputDir); !os.IsNotExist(statErr) {
		t.Fatalf("output directory was created for missing input: %v", statErr)
	}
}

func TestRunRejectsMalformedInputBeforeCreatingOutput(t *testing.T) {
	root := t.TempDir()
	inputPath := writeTestInput(t, root, "not;a;valid;row\n")
	outputDir := filepath.Join(root, "output")
	err := run(runOptions{inputPath: inputPath, outputDir: outputDir, seed: 1})
	if err == nil || !strings.Contains(err.Error(), "invalid input CSV") {
		t.Fatalf("malformed input error = %v", err)
	}
	if _, statErr := os.Stat(outputDir); !os.IsNotExist(statErr) {
		t.Fatalf("output directory was created for malformed input: %v", statErr)
	}
}

func TestRunDoesNotOverwriteExistingOutput(t *testing.T) {
	root := t.TempDir()
	inputPath := writeTestInput(t, root, testInput)
	outputDir := filepath.Join(root, "existing")
	if err := os.Mkdir(outputDir, 0o755); err != nil {
		t.Fatalf("create existing output: %v", err)
	}
	sentinelPath := filepath.Join(outputDir, "sentinel.txt")
	const sentinel = "keep this result"
	if err := os.WriteFile(sentinelPath, []byte(sentinel), 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	err := run(runOptions{inputPath: inputPath, outputDir: outputDir, seed: 1})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("existing output error = %v", err)
	}
	content, readErr := os.ReadFile(sentinelPath)
	if readErr != nil {
		t.Fatalf("read sentinel after rejected run: %v", readErr)
	}
	if string(content) != sentinel {
		t.Fatalf("existing output was changed: got %q", content)
	}
}

func TestParseRunOptionsRejectsUnknownFlag(t *testing.T) {
	_, err := parseRunOptions([]string{"-unknown"})
	if err == nil || !strings.Contains(err.Error(), "invalid command line") {
		t.Fatalf("unknown flag error = %v", err)
	}
}

func TestParseRunOptionsRejectsInvalidSeed(t *testing.T) {
	_, err := parseRunOptions([]string{"-output-dir", "output", "-seed", "not-an-integer"})
	if err == nil || !strings.Contains(err.Error(), "invalid command line") {
		t.Fatalf("invalid seed error = %v", err)
	}
}
