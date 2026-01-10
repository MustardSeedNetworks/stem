//go:build ignore

package main

import (
	"fmt"
	reflectorDP "github.com/krisarmstrong/stem/internal/reflector/dataplane"
)

func main() {
	// Try to create a dataplane instance
	dp := &reflectorDP.Dataplane{}
	fmt.Printf("dp: %+v\n", dp)
	fmt.Printf("IsRunning: %v\n", dp.IsRunning())
	fmt.Printf("GetStats: %+v\n", dp.GetStats())
}
