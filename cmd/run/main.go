package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"ising_project/ising"
)

func main() {
	var sim *ising.Simulator
	var currentL, currentCopies int

	inputFile, err := os.Open("input.csv")
	if err != nil {
		log.Fatalf("cannot open input.csv: %v", err)
	}
	defer inputFile.Close()

	reader := csv.NewReader(inputFile)
	reader.Comma = ';'

	outputFile, err := os.Create("results.csv")
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

		if len(record) != 13 {
			log.Fatalf("expected 13 fields in input.csv, got %d: %v", len(record), record)
		}

		if rowIndex == 0 && strings.EqualFold(strings.TrimSpace(record[0]), "L") {
			continue
		}

		L, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatalf("invalid L value %q: %v", record[0], err)
		}

		J1, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Fatalf("invalid J1 value %q: %v", record[1], err)
		}

		_, err = strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Fatalf("invalid J2 value %q: %v", record[2], err)
		}
		_, err = strconv.ParseFloat(record[3], 64)
		if err != nil {
			log.Fatalf("invalid J3 value %q: %v", record[3], err)
		}
		_, err = strconv.ParseFloat(record[4], 64)
		if err != nil {
			log.Fatalf("invalid J4 value %q: %v", record[4], err)
		}
		_, err = strconv.ParseFloat(record[5], 64)
		if err != nil {
			log.Fatalf("invalid J5 value %q: %v", record[5], err)
		}
		_, err = strconv.ParseFloat(record[6], 64)
		if err != nil {
			log.Fatalf("invalid J6 value %q: %v", record[6], err)
		}

		copies, err := strconv.Atoi(record[7])
		if err != nil {
			log.Fatalf("invalid copies value %q: %v", record[7], err)
		}

		h, err := strconv.ParseFloat(record[8], 64)
		if err != nil {
			log.Fatalf("invalid h value %q: %v", record[8], err)
		}

		T, err := strconv.ParseFloat(record[9], 64)
		if err != nil {
			log.Fatalf("invalid T value %q: %v", record[9], err)
		}

		aSteps, err := strconv.Atoi(record[10])
		if err != nil {
			log.Fatalf("invalid aSteps value %q: %v", record[10], err)
		}

		mSteps, err := strconv.Atoi(record[11])
		if err != nil {
			log.Fatalf("invalid mSteps value %q: %v", record[11], err)
		}

		saveVal, err := strconv.Atoi(record[12])
		if err != nil {
			log.Fatalf("invalid save value %q: %v", record[12], err)
		}
		save := saveVal != 0

		if sim == nil || !save || L != currentL || copies != currentCopies {
			var errSim error
			sim, errSim = ising.NewSimulator(L, copies)
			if errSim != nil {
				log.Fatalf("cannot create simulator: %v", errSim)
			}
			currentL = L
			currentCopies = copies
		}

		last, err := sim.Run(J1, h, T, aSteps, mSteps)
		if err != nil {
			log.Fatalf("simulation failed for L=%d, J=%.3f, h=%.3f: %v", L, J1, h, err)
		}

		//inputParams := strings.Join(record, ";")
		outRecord := []string{
			//inputParams,
			fmt.Sprintf("%f", last.E),
			fmt.Sprintf("%f", last.E2),
			fmt.Sprintf("%f", last.Mtot),
			fmt.Sprintf("%f", last.M2),
			fmt.Sprintf("%f", last.Afm),
			fmt.Sprintf("%f", last.Afm2),
		}

		if err := writer.Write(outRecord); err != nil {
			log.Fatalf("cannot write to results.csv: %v", err)
		}
	}
}
