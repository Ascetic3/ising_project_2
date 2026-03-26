package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"ising_project/internal/csvio"
	"ising_project/ising"
)

func main() {
	var sim *ising.Simulator
	var currentL, currentCopies int

	inputFile, err := os.Open("data/input/input.csv")
	if err != nil {
		log.Fatalf("cannot open input.csv: %v", err)
	}
	defer inputFile.Close()

	reader := csv.NewReader(inputFile)
	reader.Comma = ';'

	outputFile, err := os.Create("data/output/output.csv")
	if err != nil {
		log.Fatalf("cannot create results.csv: %v", err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	writer.Comma = ';'
	defer writer.Flush()

	for rowIndex := 0; ; rowIndex++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error reading input.csv: %v", err)
		}

		params, skip, err := csvio.ParseRecord(record, rowIndex)
		if err != nil {
			log.Fatalf("invalid input row: %v", err)
		}
		if skip {
			continue
		}

		if sim == nil || !params.Save || params.L != currentL || params.Copies != currentCopies {
			var errSim error
			sim, errSim = ising.NewSimulator(params.L, params.Copies)
			if errSim != nil {
				log.Fatalf("cannot create simulator: %v", errSim)
			}
			currentL = params.L
			currentCopies = params.Copies
		}

		last, err := sim.Run(params.J1, params.J2, params.H, params.T, params.ASteps, params.MSteps)
		if err != nil {
			log.Fatalf("simulation failed for L=%d, J1=%.3f, J2=%.3f, h=%.3f: %v", params.L, params.J1, params.J2, params.H, err)
		}

		outRecord := append(record,
			fmt.Sprintf("%f", last.E),
			fmt.Sprintf("%f", last.E2),
			fmt.Sprintf("%f", last.Mtot),
			fmt.Sprintf("%f", last.M2),
			fmt.Sprintf("%f", last.Afm),
			fmt.Sprintf("%f", last.Afm2),
		)

		if err := writer.Write(outRecord); err != nil {
			log.Fatalf("cannot write to results.csv: %v", err)
		}
	}
}
