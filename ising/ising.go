package ising

import (
	"fmt"
	"math"
	"math/rand"
)

type Params struct {
	L      int
	T1     float64
	T2     float64
	Tcount int

	ASteps int
	MSteps int
	Copies int
}

type ResultRow struct {
	T float64

	E    float64
	E2   float64
	Mtot float64
	M2   float64

	Afm  float64
	Afm2 float64
}

type array2d [][]int

func pbc(x, L int) int {
	if x < 0 {
		return x + L
	}
	return x % L
}

func calcParameters(lattice array2d, L int, J1, J2, h float64, energy, moment, afm *float64) {
	*energy = 0
	*moment = 0
	*afm = 0
	for x := 0; x < L; x++ {
		for y := 0; y < L; y++ {
			S := lattice[x][y]
			Sr := lattice[pbc(x+1, L)][y]
			Sb := lattice[x][pbc(y+1, L)]
			Sd1 := lattice[pbc(x+1, L)][pbc(y+1, L)]
			Sd2 := lattice[pbc(x-1, L)][pbc(y+1, L)]
			*energy += -J1 * float64(S) * float64(Sr)
			*energy += -J1 * float64(S) * float64(Sb)
			*energy += -J2 * float64(S) * float64(Sd1)
			*energy += -J2 * float64(S) * float64(Sd2)
			*energy += -h * float64(S)
			*moment += float64(S)

			sign := 1.0
			if (x+y)%2 != 0 {
				sign = -1.0
			}
			*afm += sign * float64(S)
		}
	}
}

func mcStep(lattice array2d, L int, J1, J2, h, T float64, x, y int) {
	S0 := lattice[x][y]
	S1 := -S0
	Sr := lattice[pbc(x+1, L)][y]
	Sb := lattice[x][pbc(y+1, L)]
	Sl := lattice[pbc(x-1, L)][y]
	St := lattice[x][pbc(y-1, L)]
	Sd1 := lattice[pbc(x+1, L)][pbc(y+1, L)]
	Sd2 := lattice[pbc(x-1, L)][pbc(y+1, L)]
	Sd3 := lattice[pbc(x+1, L)][pbc(y-1, L)]
	Sd4 := lattice[pbc(x-1, L)][pbc(y-1, L)]
	nnSum := Sl + Sr + St + Sb
	nnnSum := Sd1 + Sd2 + Sd3 + Sd4
	dE := float64(S1-S0) * (-h - J1*float64(nnSum) - J2*float64(nnnSum))
	if rand.Float64() < math.Exp(-dE/T) {
		lattice[x][y] = S1
	}
}

func nextStep(lattice array2d, L int, J1, J2, h, T float64) {
	N := L * L
	for i := 0; i < N; i++ {
		x := rand.Intn(L)
		y := rand.Intn(L)
		mcStep(lattice, L, J1, J2, h, T, x, y)
	}
}

type Simulator struct {
	L        int
	Copies   int
	lattices []array2d
}

func NewSimulator(L, copies int) (*Simulator, error) {
	if L <= 0 {
		return nil, fmt.Errorf("L must be > 0")
	}
	if copies <= 0 {
		return nil, fmt.Errorf("Copies must be > 0")
	}

	s := &Simulator{
		L:      L,
		Copies: copies,
	}
	s.ResetFerromagnetic()
	return s, nil
}

// ResetFerromagnetic переинициализирует все решётки ферромагнитным
// состоянием (все спины = +1) для текущих L и числа копий.
func (s *Simulator) ResetFerromagnetic() {
	L := s.L
	copies := s.Copies

	lattices := make([]array2d, copies)
	for k := 0; k < copies; k++ {
		lattice := make(array2d, 0, L)
		for i := 0; i < L; i++ {
			row := make([]int, L)
			for j := 0; j < L; j++ {
				row[j] = 1
			}
			lattice = append(lattice, row)
		}
		lattices[k] = lattice
	}
	s.lattices = lattices
}

func (s *Simulator) Run(J1, J2, h, T float64, aSteps, mSteps int) (ResultRow, error) {
	if aSteps <= 0 {
		return ResultRow{}, fmt.Errorf("ASteps must be > 0")
	}
	if mSteps <= 0 {
		return ResultRow{}, fmt.Errorf("MSteps must be > 0")
	}

	L := s.L
	copies := s.Copies

	//N := L * L

	E := 0.0
	E2 := 0.0
	M := 0.0
	M2 := 0.0
	Afm := 0.0
	Afm2 := 0.0

	for copyIdx := 0; copyIdx < copies; copyIdx++ {
		lattice := s.lattices[copyIdx]

		for sIdx := 0; sIdx < aSteps; sIdx++ {
			nextStep(lattice, L, J1, J2, h, T)
		}

		// Измерения.
		for sIdx := 0; sIdx < mSteps; sIdx++ {
			nextStep(lattice, L, J1, J2, h, T)

			energy := 0.0
			moment := 0.0
			afm := 0.0
			calcParameters(lattice, L, J1, J2, h, &energy, &moment, &afm)

			E += energy / float64(mSteps) / float64(copies)
			E2 += energy * energy / float64(mSteps) / float64(copies)
			M += math.Abs(moment) / float64(mSteps) / float64(copies)
			M2 += moment * moment / float64(mSteps) / float64(copies)

			Afm += math.Abs(afm) / float64(mSteps) / float64(copies)
			Afm2 += afm * afm / float64(mSteps) / float64(copies)
		}
	}

	return ResultRow{
		T:    T,
		E:    E,
		E2:   E2,
		Mtot: M,
		M2:   M2,

		Afm:  Afm,
		Afm2: Afm2,
	}, nil
}
