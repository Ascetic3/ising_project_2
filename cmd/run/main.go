package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"ising_project/internal/csvio"
	"ising_project/ising"
)

type runOptions struct {
	inputPath  string
	outputDir  string
	seed       int64
	saveImages bool
}

type inputRow struct {
	record []string
	params csvio.Params
}

type metadataPoint struct {
	L            int     `json:"L"`
	Copies       int     `json:"copies"`
	J1           float64 `json:"J1"`
	J2           float64 `json:"J2"`
	J3           float64 `json:"J3"`
	J4           float64 `json:"J4"`
	J5           float64 `json:"J5"`
	J6           float64 `json:"J6"`
	H            float64 `json:"h"`
	Temperature  float64 `json:"temperature"`
	AnnealSteps  int     `json:"aSteps"`
	MeasureSteps int     `json:"mSteps"`
	Save         bool    `json:"save"`
}

type metadataCLI struct {
	InputPath  string `json:"input"`
	OutputDir  string `json:"output_dir"`
	Seed       int64  `json:"seed"`
	SaveImages bool   `json:"save_images"`
}

type runMetadata struct {
	Seed         int64           `json:"seed"`
	Temperatures []float64       `json:"temperatures"`
	InputFile    string          `json:"input_file"`
	GoVersion    string          `json:"go_version"`
	StartedAt    string          `json:"started_at"`
	CLI          metadataCLI     `json:"cli"`
	Points       []metadataPoint `json:"points"`
}

const pointSeedStep int64 = 1_000_000_007

func main() {
	if err := runCLI(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCLI(args []string) error {
	options, err := parseRunOptions(args)
	if err != nil {
		return err
	}
	return run(options)
}

func parseRunOptions(args []string) (runOptions, error) {
	options := runOptions{}
	flags := flag.NewFlagSet("ising-run", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.inputPath, "input", filepath.Join("data", "input", "input.csv"), "input CSV path")
	flags.StringVar(&options.outputDir, "output-dir", "", "new output directory")
	flags.Int64Var(&options.seed, "seed", time.Now().UnixNano(), "base random seed")
	flags.BoolVar(&options.saveImages, "save-images", false, "save a lattice PNG for every point")
	if err := flags.Parse(args); err != nil {
		return runOptions{}, fmt.Errorf("invalid command line: %w", err)
	}
	if flags.NArg() != 0 {
		return runOptions{}, fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}

	options.inputPath = strings.TrimSpace(options.inputPath)
	options.outputDir = strings.TrimSpace(options.outputDir)
	if options.inputPath == "" {
		return runOptions{}, fmt.Errorf("-input must not be empty")
	}
	if options.outputDir == "" {
		return runOptions{}, fmt.Errorf("-output-dir is required")
	}
	return options, nil
}

func derivePointSeed(baseSeed int64, pointIndex int) int64 {
	return baseSeed + int64(pointIndex)*pointSeedStep
}

func readInputRows(path string) (string, []inputRow, error) {
	absolutePath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", nil, fmt.Errorf("resolve input path: %w", err)
	}
	inputFile, err := os.Open(absolutePath)
	if err != nil {
		return "", nil, fmt.Errorf("cannot open input CSV %q: %w", absolutePath, err)
	}
	defer inputFile.Close()

	reader := csv.NewReader(inputFile)
	reader.Comma = ';'
	rows := make([]inputRow, 0)
	for rowIndex := 0; ; rowIndex++ {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return "", nil, fmt.Errorf("read input CSV row %d: %w", rowIndex+1, readErr)
		}
		params, skip, parseErr := csvio.ParseRecord(record, rowIndex)
		if parseErr != nil {
			return "", nil, fmt.Errorf("invalid input CSV row %d: %w", rowIndex+1, parseErr)
		}
		if skip {
			continue
		}
		if params.L <= 0 {
			return "", nil, fmt.Errorf("invalid input CSV row %d: L must be > 0", rowIndex+1)
		}
		if params.T <= 0 {
			return "", nil, fmt.Errorf("invalid input CSV row %d: T must be > 0", rowIndex+1)
		}
		if params.ASteps <= 0 || params.MSteps <= 0 {
			return "", nil, fmt.Errorf("invalid input CSV row %d: aSteps and mSteps must be > 0", rowIndex+1)
		}
		rows = append(rows, inputRow{record: append([]string(nil), record...), params: params})
	}
	if len(rows) == 0 {
		return "", nil, fmt.Errorf("input CSV contains no calculation points")
	}
	return absolutePath, rows, nil
}

func createExclusiveOutputDir(path string) (string, error) {
	absolutePath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("resolve output directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return "", fmt.Errorf("cannot create output parent directory: %w", err)
	}
	if err := os.Mkdir(absolutePath, 0o755); err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("output directory already exists: %s", absolutePath)
		}
		return "", fmt.Errorf("cannot create output directory %q: %w", absolutePath, err)
	}
	return absolutePath, nil
}

func copyInputFile(source, outputDir string) error {
	content, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read input for output copy: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "input.csv"), content, 0o644); err != nil {
		return fmt.Errorf("write input copy: %w", err)
	}
	return nil
}

func createCSV(outputDir, name string) (*os.File, *csv.Writer, error) {
	file, err := os.OpenFile(filepath.Join(outputDir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("create %s: %w", name, err)
	}
	writer := csv.NewWriter(file)
	writer.Comma = ';'
	return file, writer, nil
}

func flushCSV(name string, writer *csv.Writer) error {
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush %s: %w", name, err)
	}
	return nil
}

func resultRecord(params csvio.Params, result ising.ResultRow) []string {
	N := float64(params.L * params.L)
	C := math.Abs(result.E2-result.E*result.E) / (params.T * params.T * N)
	kappa := math.Abs(result.M2-result.Mtot*result.Mtot) / (params.T * N)
	afKappa := math.Abs(result.Afm2-result.Afm*result.Afm) / (params.T * N)
	return []string{
		strconv.FormatFloat(params.T, 'f', -1, 64),
		strconv.FormatFloat(result.E/N, 'g', 17, 64),
		strconv.FormatFloat(result.Mtot/N, 'g', 17, 64),
		strconv.FormatFloat(result.Afm/N, 'g', 17, 64),
		strconv.FormatFloat(C, 'g', 17, 64),
		strconv.FormatFloat(kappa, 'g', 17, 64),
		strconv.FormatFloat(afKappa, 'g', 17, 64),
	}
}

func latticeImageFilename(outputDir string, pointIndex int, params csvio.Params) string {
	temperature := strconv.FormatFloat(params.T, 'f', -1, 64)
	name := fmt.Sprintf("lattice_point%03d_L%d_T%s.png", pointIndex+1, params.L, temperature)
	return filepath.Join(outputDir, "images", name)
}

func saveLatticePNG(spins [][]int, filename string) error {
	if len(spins) == 0 || len(spins[0]) == 0 {
		return fmt.Errorf("lattice is empty")
	}
	width := len(spins)
	height := len(spins[0])
	for x := range spins {
		if len(spins[x]) != height {
			return fmt.Errorf("lattice row %d has inconsistent size", x)
		}
	}

	const cellSize = 10
	img := image.NewRGBA(image.Rect(0, 0, width*cellSize, height*cellSize))
	positive := color.RGBA{R: 220, G: 45, B: 45, A: 255}
	negative := color.RGBA{R: 35, G: 80, B: 200, A: 255}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			pixel := negative
			if spins[x][y] == 1 {
				pixel = positive
			}
			for px := x * cellSize; px < (x+1)*cellSize; px++ {
				for py := y * cellSize; py < (y+1)*cellSize; py++ {
					img.Set(px, py, pixel)
				}
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return fmt.Errorf("create image directory: %w", err)
	}
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create lattice PNG: %w", err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("encode lattice PNG: %w", err)
	}
	return nil
}

func buildMetadata(options runOptions, inputPath, outputDir string, rows []inputRow, startedAt time.Time) runMetadata {
	safeRelative := func(path string) (string, bool) {
		cleaned := filepath.Clean(path)
		parentPrefix := ".." + string(os.PathSeparator)
		if cleaned == "." || cleaned == ".." || filepath.IsAbs(cleaned) ||
			filepath.VolumeName(cleaned) != "" || strings.HasPrefix(cleaned, parentPrefix) {
			return "", false
		}
		return filepath.ToSlash(cleaned), true
	}

	safeInputPath := filepath.Base(inputPath)
	if workingDir, err := os.Getwd(); err == nil {
		if relativeInput, err := filepath.Rel(workingDir, inputPath); err == nil {
			if relativeInput, ok := safeRelative(relativeInput); ok {
				safeInputPath = relativeInput
			}
		}
	}
	safeOutputDir := filepath.Base(outputDir)
	if relativeOutput, ok := safeRelative(options.outputDir); ok {
		safeOutputDir = relativeOutput
	}

	points := make([]metadataPoint, len(rows))
	temperatures := make([]float64, len(rows))
	for index, row := range rows {
		params := row.params
		temperatures[index] = params.T
		points[index] = metadataPoint{
			L: params.L, Copies: params.Copies,
			J1: params.J1, J2: params.J2, J3: params.J3,
			J4: params.J4, J5: params.J5, J6: params.J6,
			H: params.H, Temperature: params.T,
			AnnealSteps: params.ASteps, MeasureSteps: params.MSteps,
			Save: params.Save,
		}
	}
	return runMetadata{
		Seed:         options.seed,
		Temperatures: temperatures,
		InputFile:    safeInputPath,
		GoVersion:    runtime.Version(),
		StartedAt:    startedAt.Format(time.RFC3339Nano),
		CLI: metadataCLI{
			InputPath: safeInputPath, OutputDir: safeOutputDir,
			Seed: options.seed, SaveImages: options.saveImages,
		},
		Points: points,
	}
}

func writeMetadata(outputDir string, metadata runMetadata) error {
	content, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("encode run metadata: %w", err)
	}
	content = append(content, '\n')
	if err := os.WriteFile(filepath.Join(outputDir, "run_metadata.json"), content, 0o644); err != nil {
		return fmt.Errorf("write run metadata: %w", err)
	}
	return nil
}

func run(options runOptions) error {
	startedAt := time.Now().UTC()
	inputPath, rows, err := readInputRows(options.inputPath)
	if err != nil {
		return err
	}
	outputDir, err := createExclusiveOutputDir(options.outputDir)
	if err != nil {
		return err
	}
	if err := copyInputFile(inputPath, outputDir); err != nil {
		return err
	}

	outputFile, outputWriter, err := createCSV(outputDir, "output.csv")
	if err != nil {
		return err
	}
	defer outputFile.Close()
	resultFile, resultWriter, err := createCSV(outputDir, "result.csv")
	if err != nil {
		return err
	}
	defer resultFile.Close()
	diagnosticsFile, diagnosticsWriter, err := createCSV(outputDir, "diagnostics.csv")
	if err != nil {
		return err
	}
	defer diagnosticsFile.Close()
	if err := diagnosticsWriter.Write([]string{"point", "T", "point_seed"}); err != nil {
		return fmt.Errorf("write diagnostics header: %w", err)
	}

	var simulator *ising.Simulator
	currentL, currentCopies := 0, 0
	for pointIndex, row := range rows {
		params := row.params
		if simulator == nil || params.L != currentL || params.Copies != currentCopies {
			simulator, err = ising.NewSimulator(params.L, params.Copies)
			if err != nil {
				return fmt.Errorf("create simulator for point %d: %w", pointIndex+1, err)
			}
			currentL, currentCopies = params.L, params.Copies
		} else if !params.Save {
			simulator.ResetFerromagnetic()
		}

		pointSeed := derivePointSeed(options.seed, pointIndex)
		result, err := simulator.RunWithSeed(
			params.J1, params.J2, params.J3, params.J4,
			params.J5, params.J6, params.H, params.T,
			params.ASteps, params.MSteps, pointSeed,
		)
		if err != nil {
			return fmt.Errorf("simulate point %d: %w", pointIndex+1, err)
		}

		outputRecord := append(append([]string(nil), row.record...),
			strconv.FormatFloat(result.E, 'g', 17, 64),
			strconv.FormatFloat(result.E2, 'g', 17, 64),
			strconv.FormatFloat(result.Mtot, 'g', 17, 64),
			strconv.FormatFloat(result.M2, 'g', 17, 64),
			strconv.FormatFloat(result.Afm, 'g', 17, 64),
			strconv.FormatFloat(result.Afm2, 'g', 17, 64),
		)
		if err := outputWriter.Write(outputRecord); err != nil {
			return fmt.Errorf("write output.csv point %d: %w", pointIndex+1, err)
		}
		if err := resultWriter.Write(resultRecord(params, result)); err != nil {
			return fmt.Errorf("write result.csv point %d: %w", pointIndex+1, err)
		}
		if err := diagnosticsWriter.Write([]string{
			strconv.Itoa(pointIndex + 1),
			strconv.FormatFloat(params.T, 'f', -1, 64),
			strconv.FormatInt(pointSeed, 10),
		}); err != nil {
			return fmt.Errorf("write diagnostics.csv point %d: %w", pointIndex+1, err)
		}

		if options.saveImages {
			snapshot, err := simulator.LatticeSnapshot(0)
			if err != nil {
				return fmt.Errorf("snapshot point %d: %w", pointIndex+1, err)
			}
			if err := saveLatticePNG(snapshot, latticeImageFilename(outputDir, pointIndex, params)); err != nil {
				return fmt.Errorf("save image for point %d: %w", pointIndex+1, err)
			}
		}
	}

	for name, writer := range map[string]*csv.Writer{
		"output.csv": outputWriter, "result.csv": resultWriter, "diagnostics.csv": diagnosticsWriter,
	} {
		if err := flushCSV(name, writer); err != nil {
			return err
		}
	}
	metadata := buildMetadata(options, inputPath, outputDir, rows, startedAt)
	if err := writeMetadata(outputDir, metadata); err != nil {
		return err
	}

	fmt.Printf("Simulation completed: %d points\n", len(rows))
	fmt.Printf("Seed: %d\n", options.seed)
	fmt.Printf("Output directory: %s\n", outputDir)
	return nil
}
