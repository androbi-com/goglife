package main

import (
	"testing"
)

func BenchmarkLife1(b *testing.B) { life(10, 8, b.N, false, false) }
func BenchmarkLife2(b *testing.B) { life(100, 80, b.N, false, false) }
func BenchmarkLife3(b *testing.B) { life(1000, 800, b.N, false, false) }