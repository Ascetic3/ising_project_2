package csvio

import (
    "fmt"
    "strconv"
    "strings"
)

type Params struct {
    L      int
    J1     float64
    J2     float64
    J3     float64
    J4     float64
    J5     float64
    J6     float64
    Copies int
    H      float64
    T      float64
    ASteps int
    MSteps int
    Save   bool
}

func ParseRecord(record []string, rowIndex int) (Params, bool, error) {
    // Format:
    // L;J1;J2;J3;J4;J5;J6;copies;h;T;aSteps;mSteps;save
    if len(record) != 13 {
        return Params{}, false, fmt.Errorf("expected 13 fields in input.csv, got %d: %v", len(record), record)
    }

    if rowIndex == 0 && strings.EqualFold(strings.TrimSpace(record[0]), "L") {
        return Params{}, true, nil
    }

    L, err := strconv.Atoi(record[0])
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid L value %q: %w", record[0], err)
    }

    J1, err := strconv.ParseFloat(record[1], 64)
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid J1 value %q: %w", record[1], err)
    }

    J2, err := strconv.ParseFloat(record[2], 64)
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid J2 value %q: %w", record[2], err)
    }

    J3, err := strconv.ParseFloat(record[3], 64)
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid J3 value %q: %w", record[3], err)
    }
    J4, err := strconv.ParseFloat(record[4], 64)
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid J4 value %q: %w", record[4], err)
    }
    J5, err := strconv.ParseFloat(record[5], 64)
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid J5 value %q: %w", record[5], err)
    }
    J6, err := strconv.ParseFloat(record[6], 64)
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid J6 value %q: %w", record[6], err)
    }

    copies, err := strconv.Atoi(record[7])
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid copies value %q: %w", record[7], err)
    }
    if copies < 1 {
        return Params{}, false, fmt.Errorf("copies must be >= 1, got %d", copies)
    }

    h, err := strconv.ParseFloat(record[8], 64)
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid h value %q: %w", record[8], err)
    }

    T, err := strconv.ParseFloat(record[9], 64)
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid T value %q: %w", record[9], err)
    }

    aSteps, err := strconv.Atoi(record[10])
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid aSteps value %q: %w", record[10], err)
    }

    mSteps, err := strconv.Atoi(record[11])
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid mSteps value %q: %w", record[11], err)
    }

    saveVal, err := strconv.Atoi(record[12])
    if err != nil {
        return Params{}, false, fmt.Errorf("invalid save value %q: %w", record[12], err)
    }

    return Params{
        L:      L,
        J1:     J1,
        J2:     J2,
        J3:     J3,
        J4:     J4,
        J5:     J5,
        J6:     J6,
        Copies: copies,
        H:      h,
        T:      T,
        ASteps: aSteps,
        MSteps: mSteps,
        Save:   saveVal != 0,
    }, false, nil
}
