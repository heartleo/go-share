package memory_model

import (
	"fmt"
	"sync"
)

func OutOfOrder() {
	var N int
	for {
		N++
		var x, y, r1, r2 int
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			x = 1
			r1 = y
			fmt.Println("g1:", x, r1)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			y = 1
			r2 = x
			fmt.Println("g2:", y, r2)
		}()

		wg.Wait()

		if r1 == 0 && r2 == 0 {
			fmt.Println("N:", N)
			break
		}
	}
}
