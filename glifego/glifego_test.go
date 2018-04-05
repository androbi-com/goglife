package main

import (
	"fmt"
	"testing"
)

func BenchmarkSimulate1(b *testing.B) { life(100, 80, b.N, 2, false, false) }
func BenchmarkSimulate2(b *testing.B) { life(100, 80, b.N, 4, false, false) }
func BenchmarkSimulate3(b *testing.B) { life(1000, 800, b.N, 2, false, false) }
func BenchmarkSimulate4(b *testing.B) { life(1000, 800, b.N, 4, false, false) }

func TestUniverse_Make(t *testing.T) {
	
	t.Run("test1", func(t *testing.T) {
		var u1, u2 Universe;
		u1.Make(3,2)
		u2.Make(3,2)
		u1.cells[0][0]=1
		u2.cells[0]=u1.cells[0]
		fmt.Println(u1);
		fmt.Println(u2);
		u1.cells[0][0]=2
		fmt.Println(u2);
	})
}




