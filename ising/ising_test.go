package ising

import (
	"math"
	"math/rand"
	"testing"
)

const deltaETolerance = 1e-10

type testModel struct {
	J1, J2, J3, J4, J5, J6 float64
}

func testFullEnergy(lattice array2d, L int, model testModel, h float64) float64 {
	energy, moment, afm := 0.0, 0.0, 0.0
	calcParameters(
		lattice, L,
		model.J1, model.J2, model.J3, model.J4, model.J5, model.J6,
		h, &energy, &moment, &afm,
	)
	return energy
}

func exactDeltaEForTest(lattice array2d, L int, model testModel, h float64, x, y int) (before, after, delta float64) {
	before = testFullEnergy(lattice, L, model, h)
	lattice[x][y] = -lattice[x][y]
	after = testFullEnergy(lattice, L, model, h)
	lattice[x][y] = -lattice[x][y]
	return before, after, after - before
}

func localDeltaEForTest(lattice array2d, L int, model testModel, h float64, x, y int) float64 {
	return localDeltaEForFlip(
		lattice, L,
		model.J1, model.J2, model.J3, model.J4, model.J5, model.J6,
		h, x, y,
	)
}

func assertDeltaEEqual(t *testing.T, lattice array2d, L int, model testModel, h float64, x, y int) float64 {
	t.Helper()
	spin := lattice[x][y]
	before, after, exact := exactDeltaEForTest(lattice, L, model, h, x, y)
	local := localDeltaEForTest(lattice, L, model, h, x, y)
	diff := math.Abs(local - exact)
	if lattice[x][y] != spin {
		t.Fatalf("test changed lattice irreversibly at (%d,%d): before=%d after=%d", x, y, spin, lattice[x][y])
	}
	if diff > deltaETolerance {
		class := siteClass(x, y)
		c := couplingsForSite(class, model.J1, model.J2, model.J3, model.J4, model.J5, model.J6)
		t.Fatalf(
			"delta E mismatch: L=%d x=%d y=%d class=%d spin=%+d h=%.17g couplings={up:%.17g right:%.17g down:%.17g left:%.17g dl:%.17g dr:%.17g ur:%.17g ul:%.17g} E_before=%.17g E_after=%.17g exact_dE=%.17g local_dE=%.17g diff=%.17g",
			L, x, y, class, spin, h,
			c.up, c.right, c.down, c.left, c.dl, c.dr, c.ur, c.ul,
			before, after, exact, local, diff,
		)
	}
	return diff
}

func makeTestLattice(L int, spinAt func(x, y int) int) array2d {
	lattice := make(array2d, L)
	for x := 0; x < L; x++ {
		lattice[x] = make([]int, L)
		for y := 0; y < L; y++ {
			lattice[x][y] = spinAt(x, y)
		}
	}
	return lattice
}

func randomTestLattice(L int, rng *rand.Rand) array2d {
	return makeTestLattice(L, func(_, _ int) int {
		if rng.Intn(2) == 0 {
			return -1
		}
		return 1
	})
}

func TestLocalDeltaEConsistencyRandom(t *testing.T) {
	model := testModel{J1: 0.71, J2: -1.13, J3: 1.37, J4: -0.89, J5: 0.53, J6: -1.61}
	fields := []float64{0, 0.27}
	rng := rand.New(rand.NewSource(20260731))
	seenClassSpins := [4][2]bool{}
	checks := 0
	maxDiff := 0.0

	for _, L := range []int{4, 8} {
		for _, h := range fields {
			for configuration := 0; configuration < 8; configuration++ {
				lattice := randomTestLattice(L, rng)
				for x := 0; x < L; x++ {
					for y := 0; y < L; y++ {
						spinIndex := 0
						if lattice[x][y] == 1 {
							spinIndex = 1
						}
						seenClassSpins[siteClass(x, y)][spinIndex] = true
						diff := assertDeltaEEqual(t, lattice, L, model, h, x, y)
						if diff > maxDiff {
							maxDiff = diff
						}
						checks++
					}
				}
			}
		}
	}

	if checks < 1000 {
		t.Fatalf("performed %d checks, want at least 1000", checks)
	}
	for class, seenSpins := range seenClassSpins {
		if !seenSpins[0] || !seenSpins[1] {
			t.Fatalf("siteClass %d did not check both flip directions: seen[-1,+1]=%v", class, seenSpins)
		}
	}
	t.Logf("delta E random checks=%d max_abs_diff=%.17g", checks, maxDiff)
}

func TestLocalDeltaEConsistencyManualStates(t *testing.T) {
	model := testModel{J1: 0.71, J2: -1.13, J3: 1.37, J4: -0.89, J5: 0.53, J6: -1.61}
	pattern := [][]int{
		{1, -1, -1, 1},
		{-1, -1, 1, 1},
		{1, 1, -1, 1},
		{-1, 1, 1, -1},
	}

	tests := []struct {
		name    string
		lattice array2d
		x, y    int
		h       float64
	}{
		{
			name:    "all spins positive",
			lattice: makeTestLattice(4, func(_, _ int) int { return 1 }),
			x:       0, y: 0, h: 0,
		},
		{
			name:    "all spins negative",
			lattice: makeTestLattice(4, func(_, _ int) int { return -1 }),
			x:       1, y: 0, h: 0.27,
		},
		{
			name: "checkerboard",
			lattice: makeTestLattice(4, func(x, y int) int {
				if (x+y)%2 == 0 {
					return 1
				}
				return -1
			}),
			x: 0, y: 1, h: 0,
		},
		{
			name: "single defect",
			lattice: makeTestLattice(4, func(x, y int) int {
				if x == 2 && y == 2 {
					return -1
				}
				return 1
			}),
			x: 1, y: 1, h: 0.27,
		},
	}

	for class, coordinate := range [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}} {
		x, y := coordinate[0], coordinate[1]
		tests = append(tests, struct {
			name    string
			lattice array2d
			x, y    int
			h       float64
		}{
			name: "explicit site class " + string(rune('0'+class)),
			lattice: makeTestLattice(4, func(x, y int) int {
				return pattern[x][y]
			}),
			x: x, y: y, h: 0.27,
		})
	}

	maxDiff := 0.0
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diff := assertDeltaEEqual(t, test.lattice, len(test.lattice), model, test.h, test.x, test.y)
			if diff > maxDiff {
				maxDiff = diff
			}
		})
	}
	t.Logf("delta E manual states=%d max_abs_diff=%.17g", len(tests), maxDiff)
}
