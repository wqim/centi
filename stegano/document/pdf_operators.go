package document
import (
)

func SetupPdfOperators( bitsPerOperand int ) []IOperator {
	// just setup all the supported operators...
	return []IOperator{
		NewOperator("c", 6, 6, bitsPerOperand, 0.05, []float64{1, 1, 1, 1, 1, 1} ),
		NewOperator("v", 4, 4, bitsPerOperand, 0.05, []float64{1, 1, 1, 1}),
		NewOperator("y", 4, 4, bitsPerOperand, 0.05, []float64{1, 1, 1, 1}),
		NewOperator("l", 2, 2, bitsPerOperand, 0.05, []float64{0.05, 0.05}),
		NewOperator("m", 2, 2, bitsPerOperand, 0.05, []float64{0.05, 0.05}),
		NewOperator("re", 4, 4, bitsPerOperand, 0.01, []float64{0.2, 0.2, 0.2, 0.2}),
		NewOperator("cm", 6, 6, bitsPerOperand, 0.001, []float64{0.1, 0.1, 0.1, 0.1, 0.05, 0.05}),
		NewOperator("i", 1, 1, bitsPerOperand, 0.05, []float64{1}),
		NewOperator("M", 1, 1, bitsPerOperand, 0.05, []float64{1}),
		NewOperator("w", 1, 1, bitsPerOperand, 0.05, []float64{1}),
		NewOperator("G", 1, 1, bitsPerOperand, 0.05, []float64{5}),
		NewOperator("g", 1, 1, bitsPerOperand, 0.05, []float64{5}),
		NewOperator("K", 4, 4, bitsPerOperand, 0.05, []float64{5, 5, 5, 5}),
		NewOperator("k", 4, 4, bitsPerOperand, 0.05, []float64{5, 5, 5, 5}),
		NewOperator("RG", 3, 3, bitsPerOperand, 0.05, []float64{5, 5, 5}),
		NewOperator("rg", 3, 3, bitsPerOperand, 0.05, []float64{5, 5, 5}),
		NewOperator("sc", 1, 4, bitsPerOperand, 0.05, []float64{5, 5, 5, 5}),
		NewOperator("SC", 1, 4, bitsPerOperand, 0.05, []float64{5, 5, 5, 5}),
		NewOperator("scn", 1, 4, bitsPerOperand, 0.05, []float64{5, 5, 5, 5}),
		NewOperator("SCN", 1, 4, bitsPerOperand, 0.05, []float64{5, 5, 5, 5}),
		NewOperator("Tc", 1, 1, bitsPerOperand, 0.05, []float64{1}),
		NewOperator("Td", 2, 2, bitsPerOperand, 0.05, []float64{2, 2}),
		NewOperator("TD", 2, 2, bitsPerOperand, 0.05, []float64{2, 2}),
		NewOperator("Tf", 1, 1, bitsPerOperand, 0.05, []float64{0.5}),
		NewOperator("TL", 1, 1, bitsPerOperand, 0.05, []float64{2}),
		NewOperator("Tm", 6, 6, bitsPerOperand, 0.001, []float64{ 0.005, 0.005, 0.005, 0.005, 2, 2 }),
		NewOperator("Ts", 1, 1, bitsPerOperand, 0.05, []float64{5}),
		NewOperator("Tw", 1, 1, bitsPerOperand, 0.05, []float64{1}),
		NewOperator("Tz", 1, 1, bitsPerOperand, 0.05, []float64{0.5}),
	
		NewTjOperator("TJ", 15, bitsPerOperand, 5.0 ),

		NewOperator("d0", 1, 1, bitsPerOperand, 0.05, []float64{1}),
		NewOperator("d1", 5, 5, bitsPerOperand, 0.05, []float64{1, 1, 1, 1, 1}),
	}
}
