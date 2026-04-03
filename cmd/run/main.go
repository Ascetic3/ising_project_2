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
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	var sim *ising.Simulator
	var currentL, currentCopies int

	inputFile, err := os.Open("data/input/input.csv")
	if err != nil {
		return fmt.Errorf("cannot open input.csv: %w", err)
	}
	defer inputFile.Close()

	reader := csv.NewReader(inputFile)
	reader.Comma = ';'

	outputFile, err := os.Create("data/output/output.csv")
	if err != nil {
		return fmt.Errorf("cannot create output.csv: %w", err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	writer.Comma = ';'
	defer func() {
		writer.Flush()
		if werr := writer.Error(); werr != nil && err == nil {
			err = fmt.Errorf("csv writer: %w", werr)
		}
	}()

	for rowIndex := 0; ; rowIndex++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading input.csv: %w", err)
		}

		params, skip, err := csvio.ParseRecord(record, rowIndex)
		if err != nil {
			return fmt.Errorf("invalid input row: %w", err)
		}
		if skip {
			continue
		}

		// ВАЖНО:
		// Мы НЕ сбрасываем решётку при смене температуры.
		// Это необходимо для корректного моделирования фазового перехода.
		// Используется конфигурация предыдущей температуры.
		if sim == nil || params.L != currentL || params.Copies != currentCopies {
			sim, err = ising.NewSimulator(params.L, params.Copies)
			if err != nil {
				return fmt.Errorf("cannot create simulator: %w", err)
			}
			currentL = params.L
			currentCopies = params.Copies
		}

		last, err := sim.Run(
			params.J1, params.J2, params.J3, params.J4, params.J5, params.J6,
			params.K, params.H, params.T,
			params.ASteps, params.MSteps,
		)
		if err != nil {
			return fmt.Errorf("simulation failed for L=%d, J=[%.3f %.3f %.3f %.3f %.3f %.3f], K=%.3f, h=%.3f: %w",
				params.L, params.J1, params.J2, params.J3, params.J4, params.J5, params.J6, params.K, params.H, err)
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
			return fmt.Errorf("cannot write to output.csv: %w", err)
		}
	}

	return nil
}
