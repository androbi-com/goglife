package main

import (
	"fmt"
	"time"
)

// Universe the domain where life happens
type Universe struct {
	cells [][]uint8 // 0 = dead 1 = alive
	nx    int
	ny    int
}

// Make domain
func (u *Universe) Make(nx int, ny int) {
	u.nx = nx
	u.ny = ny
	u.cells = make([][]uint8, u.ny)
	for i := range u.cells {
		u.cells[i] = make([]uint8, u.nx)
	}
}

// SetGliderAt i,j. A glider moves 1 cell diagonally each 4 generations
func (u *Universe) SetGliderAt(i int, j int) {
	u.cells[i][j+1] = 1
	u.cells[i + 1][j + 2] = 1
	u.cells[i + 2][j] = 1
	u.cells[i + 2][j + 1] = 1
	u.cells[i + 2][j + 2] = 1
}

// Print domain
func (u *Universe) Print() {
	for i := range u.cells {
		fmt.Println(u.cells[i])
	}
}

// CopyCellsTo copies cells from u to nu (new universe)
func (u *Universe) CopyCellsTo(nu *Universe) {
	for i := range u.cells {
		for j := range u.cells[i] {
			nu.cells[i][j] = u.cells[i][j]
		}
	}
}

// EvolveOneGenerationTo store next generation in nu
func (u *Universe) EvolveOneGenerationTo(nu *Universe) {
	for i := range u.cells {
		for j := range u.cells[i] {
			n := uint8(0)
			for i1 := i - 1; i1 <= i+1; i1++ {
				for j1 := j - 1; j1 <= j+1; j1++ {
					n += u.cells[(i1 + u.ny) % u.ny][(j1 + u.nx) % u.nx] // apply period boundary conditions
				}
			}
			n -= u.cells[i][j]
			nu.cells[i][j] = 0
			if n == 3 || (n == 2 && u.cells[i][j] > 0) {
				nu.cells[i][j] = 1
			}
		}
	}
}

// testable version of main
func life(nx int, ny int, ns int, wantPrint bool, wantTimer bool) {
	start := time.Now()
	fmt.Println("start game of life on", nx, "x", ny, ",", ns, "generations.")

	// prepare univCurrent and univNext
	var univCurrent, univNext Universe
	univCurrent.Make(nx, ny)
	univNext.Make(nx, ny)
	univCurrent.SetGliderAt(0, 0)
	if wantPrint {
		univCurrent.Print()
	}

	// iterate over ns generations
	for i := 0; i < ns; i++ {
		univCurrent.EvolveOneGenerationTo(&univNext)
		univNext.CopyCellsTo(&univCurrent)
	}

	// results
	if wantTimer {
		fmt.Println(time.Since(start))
	}
	if wantPrint {
		fmt.Println("---------------------")
		univCurrent.Print()
	}
}

// main just wraps example calls to life()
func main() {
	// small universe 10x8, 16 steps (4 diagonal moves), 1 & 2 go routines, print universe, no timer
	life(10, 8, 16, true, false)

    // bigger universe 1000x800, 10 steps, no print, timer
	life(1000, 1000, 10, false, true)
}
