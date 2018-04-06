package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/db7/barrier"
)

const offSet = 1 // this is introduced to make clearer where array access is offset due to shadow row

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
	u.cells = make([][]uint8, u.ny+2) // add two shadow rows
	for i := range u.cells {
		u.cells[i] = make([]uint8, u.nx)
	}
}

// SetGliderAt i,j. A glider moves 1 cell diagonally each 4 generations
func (u *Universe) SetGliderAt(i int, j int) {
	u.cells[i+offSet][j+1] = 1
	u.cells[i+1+offSet][j+2] = 1
	u.cells[i+2+offSet][j] = 1
	u.cells[i+2+offSet][j+1] = 1
	u.cells[i+2+offSet][j+2] = 1
}

// Print domain for go routine # ngo
func (u *Universe) Print(ngo int) {
	for i := offSet; i < u.ny+offSet; i++ {
		fmt.Println(ngo, u.cells[i])
	}
}

// CopyCellsTo copies cells from u to nu (new universe)
func (u *Universe) CopyCellsTo(nu *Universe) {
	for i := offSet; i < u.ny+offSet; i++ {
		for j := range u.cells[i] {
			nu.cells[i][j] = u.cells[i][j]
		}
	}
}

// EvolveOneGenerationTo one step in life
func (u *Universe) EvolveOneGenerationTo(nu *Universe) {
	for i := offSet; i < u.ny+offSet; i++ {
		for j := range u.cells[i] {
			n := uint8(0)
			for i1 := i - 1; i1 <= i+1; i1++ {
				for j1 := j - 1; j1 <= j+1; j1++ {
					n += u.cells[i1][(j1+u.nx)%u.nx] // shadow cells handle periodic boundary conds in y
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

// GoLife go routine that evolves it's own part of the universe
func GoLife(igo int, ngo int, nx int, ny int, ns int, wantPrint bool,
	wg *sync.WaitGroup, syncGroup *barrier.Barrier,
	leftSend, rightRecv, leftRecv, rightSend chan []uint8) {

	// each go routine has its own universe which includes 2 shadow rows
	var univCurrent, univNext Universe
	univCurrent.Make(nx, ny)
	univNext.Make(nx, ny)
	if igo == 0 { // only first go routine creates a glider
		univCurrent.SetGliderAt(0, 0)
	}

	// order output by establishing a pipeline in one of our channels
	if wantPrint {
		if (ngo == 1) {
			univCurrent.Print(igo)
		} else {
			if igo == 0 {
				univCurrent.Print(igo)
				rightSend <- nil
			} else if igo < ngo-1 {
				<-leftRecv
				univCurrent.Print(igo)
				rightSend <- nil
			} else {
				<-leftRecv
				univCurrent.Print(igo)
			}
		}
	}

	// note that we send references, once they are in place data is updated automatically
	if (ngo ==1) {
		univCurrent.cells[0] = univCurrent.cells[univCurrent.ny]
		univCurrent.cells[univCurrent.ny+1] =  univCurrent.cells[1]
	} else {
		if igo%2 == 0 { // even go-routine number
			leftSend <- univCurrent.cells[1]
			univCurrent.cells[univCurrent.ny+1] = <-rightRecv
			rightSend <- univCurrent.cells[univCurrent.ny]
			univCurrent.cells[0] = <-leftRecv
		} else { // odd go-routine number
			univCurrent.cells[univCurrent.ny+1] = <-rightRecv
			leftSend <- univCurrent.cells[1]
			univCurrent.cells[0] = <-leftRecv
			rightSend <- univCurrent.cells[univCurrent.ny]
		}
	}

	// iterate over ns generations
	for i := 0; i < ns; i++ {
		// barrier: wait for all
		syncGroup.Await(func() error {
			return nil
		})
		// copy shadow cells from neighbors in the communication ring
		univCurrent.EvolveOneGenerationTo(&univNext)
		// barrier: wait for all
		syncGroup.Await(func() error {
			return nil
		})
		univNext.CopyCellsTo(&univCurrent)
	}
	// results
	if wantPrint { // order output by establishing a pipeline in one of our channels
		if (ngo == 1) {
			fmt.Println("-----------------------")
			univCurrent.Print(igo)
		} else {
			if igo == 0 {
				fmt.Println("-----------------------")
				univCurrent.Print(igo)
				rightSend <- nil
			} else if igo < ngo-1 {
				<-leftRecv
				univCurrent.Print(igo)
				rightSend <- nil
			} else {
				<-leftRecv
				univCurrent.Print(igo)
			}
		}
	}
	wg.Done()
}

// testable version of main
func life(nx int, ny int, ns int, ngo int, wantPrint bool, wantTimer bool) {
	fmt.Println("start game of life on", nx, "x", ny, ",", ns, "generations with", ngo, "go routines")
	start := time.Now()
	// some primitive error checking
	if ngo > 1 && ngo%2 != 0 {
		fmt.Println("Error: ngo must be 1 or divisible by 2")
		os.Exit(1)
	}
	if ny%ngo != 0 {
		fmt.Println("Error: ny must be divisble by ngo")
		os.Exit(1)
	}
	// prepare communication channels for ring. Example np = 2:
	//  i        0    1     0      at the ends channel "wraps around"
	//  left:   <- * -<- * -<      send to left  (receive from right)
	//  right:  >- * ->- * ->      send to right (receive from left)
	left := make([]chan []uint8, ngo)
	right := make([]chan []uint8, ngo)
	for i := range left {
		left[i] = make(chan []uint8)
		right[i] = make(chan []uint8)
	}

	// sync tools
	var wg sync.WaitGroup
	b := barrier.New(ngo)

	// create go routines
	for i := 0; i < ngo; i++ {
		wg.Add(1)      // add go routine to wait group
		if i < ngo-1 { // leftS    rightR     leftR     rightS
			go GoLife(i, ngo, nx, ny/ngo, ns, wantPrint, &wg, b, left[i], left[i+1], right[i], right[i+1])
		} else {
			go GoLife(i, ngo, nx, ny/ngo, ns, wantPrint, &wg, b, left[ngo-1], left[0], right[ngo-1], right[0])
		}
	}

	// wait for all go routines to finish
	wg.Wait()
	if wantTimer {
		fmt.Println(time.Since(start))
	}
}

// main just wraps example calls to life()
func main() {
	// small universe 10x8, 16 steps (4 diagonal moves), 1 & 2 go routines, print universe, no timer
	life(10, 8, 16, 1, true, false)
	life(10, 8, 16, 2, true, false)
	// bigger universe 1000x800, 10 steps, 1-2-4 go-routines, no print, timer
	life(1000, 1000, 10, 1, false, true)
	life(1000, 1000, 10, 2, false, true)
	life(1000, 1000, 10, 4, false, true)
}
