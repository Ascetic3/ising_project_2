package ising

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
)

// type Params struct {
// 	L      int
// 	T1     float64
// 	T2     float64
// 	Tcount int

// 	ASteps int
// 	MSteps int
// 	Copies int
// }

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

// Couplings задаёт направленные коэффициенты обменного взаимодействия на Union Jack.
// Направления соответствуют соседям относительно узла (x,y):
//
//	up    : (x,   y-1)
//	right : (x+1, y  )
//	down  : (x,   y+1)
//	left  : (x-1, y  )
//	ur    : (x+1, y-1)
//	dr    : (x+1, y+1)
//	ul    : (x-1, y-1)
//	dl    : (x-1, y+1)
type Couplings struct {
	up, right, down, left float64
	dl, dr, ur, ul        float64
}

// siteClass возвращает класс узла в элементарной ячейке 2x2:
//
//	0 = blue  (x%2==0 && y%2==0)
//	1 = red1 (x%2==1 && y%2==0)
//	2 = red2 (x%2==0 && y%2==1)
//	3 = red1 (x%2==1 && y%2==1)
func siteClass(x, y int) int {
	xEven := x%2 == 0
	yEven := y%2 == 0
	switch {
	case xEven && yEven:
		return 0
	case !xEven && yEven:
		return 1
	case xEven && !yEven:
		return 2
	default:
		return 3
	}
}

// couplingsForSite восстанавливает направленные J для узла заданного класса,
// чтобы каждая связь имела согласованный коэффициент с обеих сторон.

func couplingsForSite(class int, J1, J2, J3, J4, J5, J6 float64) Couplings {
	switch class {
	case 0: // blue
		return Couplings{
			up: J1, right: J2, down: J3, left: J4,
			dl: J5, dr: J6, ur: J5, ul: J6,
		}
	case 1: // red1 справа-вверху
		return Couplings{
			up: J3, right: J4, down: J1, left: J2,
			dl: 0, dr: 0, ur: 0, ul: 0,
		}
	case 2: // red2 справа-вниз
		return Couplings{
			up: J3, right: J2, down: J1, left: J4,
			ur: J6, dr: J5, ul: J5, dl: J6,
		}
	default: // 3: red1 снизу-слева
		return Couplings{
			up: J3, right: J4, down: J1, left: J2,
			dl: 0, dr: 0, ur: 0, ul: 0,
		}
	}
}

func calcParameters(lattice array2d, L int, J1, J2, J3, J4, J5, J6, h float64, energy, moment, afm *float64) {
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

			// Обменное взаимодействие: учитываем только уникальные связи
			// right, down, down-right, down-left.
			c := couplingsForSite(siteClass(x, y), J1, J2, J3, J4, J5, J6)
			*energy += -c.right * float64(S) * float64(Sr)
			*energy += -c.down * float64(S) * float64(Sb)
			*energy += -c.dr * float64(S) * float64(Sd1)
			*energy += -c.dl * float64(S) * float64(Sd2)

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

func mcStep(rng *rand.Rand, lattice array2d, L int, J1, J2, J3, J4, J5, J6, h, T float64, x, y int) {
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

	// Класс узла определяет, какие направленные J использовать для восьми ближайших соседей.
	c := couplingsForSite(siteClass(x, y), J1, J2, J3, J4, J5, J6)

	// pairSum — сумма по 8 направлениям (все спины-соседи помножены на соответствующие коэффициенты).
	pairSum := c.up*float64(St) +
		c.right*float64(Sr) +
		c.down*float64(Sb) +
		c.left*float64(Sl) +
		c.dl*float64(Sd2) +
		c.dr*float64(Sd1) +
		c.ur*float64(Sd3) +
		c.ul*float64(Sd4)

	dE := 2 * float64(S0) * (pairSum + h)
	if rng.Float64() < math.Exp(-dE/T) {
		lattice[x][y] = S1
	}
}

func nextStep(rng *rand.Rand, lattice array2d, L int, J1, J2, J3, J4, J5, J6, h, T float64) {
	N := L * L
	for i := 0; i < N; i++ {
		x := rng.Intn(L)
		y := rng.Intn(L)
		mcStep(rng, lattice, L, J1, J2, J3, J4, J5, J6, h, T, x, y)
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

func (s *Simulator) Run(J1, J2, J3, J4, J5, J6, h, T float64, aSteps, mSteps int) (ResultRow, error) {
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

	type copyResult struct {
		E, E2, M, M2, Afm, Afm2 float64
	}

	const seedStep int64 = 1_000_003
	baseSeed := int64(1)
	weight := 1 / float64(mSteps*copies)

	results := make([]copyResult, copies)
	var wg sync.WaitGroup

	for copyIdx := 0; copyIdx < copies; copyIdx++ {
		wg.Add(1)
		go func(copyIdx int) {
			defer wg.Done()

			rng := rand.New(rand.NewSource(baseSeed + int64(copyIdx)*seedStep))
			lattice := s.lattices[copyIdx]
			local := copyResult{}

			for sIdx := 0; sIdx < aSteps; sIdx++ {
				nextStep(rng, lattice, L, J1, J2, J3, J4, J5, J6, h, T)
			}

			for sIdx := 0; sIdx < mSteps; sIdx++ {
				nextStep(rng, lattice, L, J1, J2, J3, J4, J5, J6, h, T)

				energy := 0.0
				moment := 0.0
				afm := 0.0
				calcParameters(lattice, L, J1, J2, J3, J4, J5, J6, h, &energy, &moment, &afm)

				local.E += energy * weight
				local.E2 += energy * energy * weight
				local.M += math.Abs(moment) * weight
				local.M2 += moment * moment * weight

				local.Afm += math.Abs(afm) * weight
				local.Afm2 += afm * afm * weight
			}

			results[copyIdx] = local
		}(copyIdx)
	}

	wg.Wait()

	for _, result := range results {
		E += result.E
		E2 += result.E2
		M += result.M
		M2 += result.M2
		Afm += result.Afm
		Afm2 += result.Afm2
	}
	fmt.Printf("L=%d: J=[%.3f %.3f %.3f %.3f %.3f %.3f], h=%.3f: E=%.3f, M=%.3f, Afm=%.3f \n",
		L, J1, J2, J3, J4, J5, J6, h, E, M, Afm)
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
