package internal

import (
	"sync"
	"testing"
)

// TestProvaRaceDescartavel injeta um data race de propósito — PR de prova do
// gate go test -race da task 0004. Nunca mergear.
func TestProvaRaceDescartavel(t *testing.T) {
	contador := 0
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			contador++
		}()
	}
	wg.Wait()
	t.Logf("contador final: %d", contador)
}
