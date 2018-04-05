# Conway's Game of Life with Go

One of my first little projects with the Go language. There is a simple and a concurrent version, which
has been inspired by the C + MPI program I use as a [basic example for domain decomposition in MPI in my HPC
course](http://androbi.com/course/html/mpi_gameoflife.md.html). 

The initial idea was to use the elegant formulation of go channels to visualize the necessary 
communications in the MPI program. When students start programming with MPI, the communication part of
a MPI program usually poses the most difficulties.

In the MPI program each process holds a 2D array of the domain data and arrays corresponding 
to the so called shadow cells are copied between processes (see link above).

However, when writing the concurrent Go version, I noted that in Go slices are the most appropiate 
data type for holding the domain data, as they can by created dynamically. Thus, the Go program
interchanges slices between go routines during communication. As slices actually are references 
to the array elements of the neighboring domains which reside in the same address space, the 
communications are only necessary once, at the beginning of the program! The shadow cells will 
automatically hold the correct values when the original cells are updated. Some care has to be 
taken in order to avoid cell updates in one go routine while in another one we are still evaluating 
the previous generation. This can be done by imposing barriers, I have used the 
package [http://github.com/db7/barrier] for this purpose.

This program was not designed for maximum efficency, probably a one dimensional slice with suitable
mapping function for 2D access would be more efficient. I have observed good speedups
in my first trial runs (this is on a 2-core i7-4600U CPU)

    start game of life on 1000 x 800 , 10 generations with 1 go routines
    1.287180352s
    start game of life on 1000 x 800 , 10 generations with 2 go routines
    820.781606ms
    start game of life on 1000 x 800 , 10 generations with 4 go routines
    548.347335ms
