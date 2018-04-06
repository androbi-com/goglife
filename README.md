# Conway's Game of Life with Go

One of my first little projects with the Go language. There is a simple and a concurrent version, which
has been inspired by the C + MPI program I use as a [basic example for domain decomposition in MPI in my HPC
course](http://androbi.com/course/html/mpi_gameoflife.md.html). 

The initial idea was to use the elegant formulation of go channels to visualize the necessary 
communications in the MPI program. When students start programming with MPI, the communication part of
a MPI program usually poses the most difficulties.

In the MPI program each process holds a 2D array of the domain data and arrays corresponding 
to the so called shadow cells are copied between processes (see link above).

However, when writing the concurrent Go version, I noted that in Go slices are the most appropriate 
data type for holding the domain data, as they can by created dynamically. Thus, the Go program
interchanges slices between go routines during communication. As slices actually are references 
to the array elements of the neighboring domains which reside in the same address space, the 
communications are only necessary once, at the beginning of the program! The shadow cells will 
automatically hold the correct values when the original cells are updated. Some care has to be 
taken in order to avoid cell updates in one go routine while in another one we are still evaluating 
the previous generation. This can be done by imposing barriers, I have used the 
package [http://github.com/db7/barrier] for this purpose.

This program was not designed for maximum efficiency, probably a one dimensional slice with suitable
mapping function for 2D access would be more efficient. In order to have some baseline to 
compare to, I ran the (C MPI version)[https://bitbucket.org/clausi/hpc/src/0de6f3566a68302f1c6235526f842e30234f2f77/src/mpi/gameoflife/?at=master] on my laptop for a 1000x1000 domain and 1000 generations with the following results

    $ mpirun -np 1 gamempi
    Process 0 of 1 is running on TECRA-Z50-A
    Done. 1000 iterations in 2.461644 secs
    
    $ mpirun -np 2 gamempi
    Process 0 of 2 is running on TECRA-Z50-A
    Process 1 of 2 is running on TECRA-Z50-A
    Done. 1000 iterations in 1.418004 secs
    
    $ mpirun -np 4 gamempi
    Process 3 of 4 is running on TECRA-Z50-A
    Process 0 of 4 is running on TECRA-Z50-A
    Process 1 of 4 is running on TECRA-Z50-A
    Process 2 of 4 is running on TECRA-Z50-A
    Done. 1000 iterations in 1.354918 secs

This is on a 2-core i7-4600U CPU, which explains why 4 processes are not much faster than 2. I 
found my simple Go version (glife.go) to be much slower, on a 1000x1000 domain 10(!) 
generations would take 2.5 secs.

So, the Go version is about a factor of 100 slower. I knew the C version would be faster but
I did not expect such a big difference. To rule out programming errors, I checked the 
implementation from golang.org (see https://golang.org/doc/play/life.go) and it 
took 2.8 secs for the same system parameters.

The C version was compiled with a vectorizing compiler with aggressive optimizations enabled. 
Removing all optimizations with -O0 incremented execution speed by a factor of 5, but there 
is still a factor of 20 left when comparing with the Go version. Fast array access is 
essential for this program, as there are only 3 integer additions and two modulo 
divisions inside the innermost main loop. So, array access, at least when implemented 
by 2D slices is not an area where Go currently shines. Fair enough to say, Go was never 
designed to be used for this kind of tasks.

When running the concurrent version of my Go implementation, I have observed a scaling 
behavior similar to the C + MPI program on my laptop: 

    start game of life on 1000 x 1000 , 10 generations with 1 go routines
    1.670505757s
    start game of life on 1000 x 1000 , 10 generations with 2 go routines
    854.053571ms
    start game of life on 1000 x 1000 , 10 generations with 4 go routines
    776.365622ms