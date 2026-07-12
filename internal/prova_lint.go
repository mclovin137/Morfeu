package internal

import "os"

// ProvaLintDescartavel viola errcheck de propósito — PR de prova do gate de
// lint da task 0004. Nunca mergear.
func ProvaLintDescartavel() {
	os.Remove("arquivo-que-nao-existe")
}
