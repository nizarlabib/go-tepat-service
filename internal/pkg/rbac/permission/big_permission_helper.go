package permission

import "math/big"

func getSetBitIndices(n *big.Int) []int {
	if n.Sign() <= 0 {
		return nil
	}
	bitLen := n.BitLen()
	if bitLen == 0 {
		return nil
	}
	// For very small numbers, use simple iteration
	if bitLen <= 64 {
		return getSetBitIndicesSimple(n, bitLen)
	}
	// Estimate density by sampling
	density := estimateBitDensity(n, bitLen)
	// Choose algorithm based on estimated density and size
	if shouldUseSparseAlgorithm(density, bitLen) {
		return getSetBitIndicesSparse(n)
	} else {
		return getSetBitIndicesDense(n, bitLen)
	}
}

func getSetBitIndicesSimple(n *big.Int, bitLen int) []int {
	var indices []int
	for i := range bitLen {
		if n.Bit(i) == 1 {
			indices = append(indices, i)
		}
	}
	return indices
}

// Optimized for sparse bit patterns (few set bits)
func getSetBitIndicesSparse(n *big.Int) []int {
	var indices []int
	temp := new(big.Int).Set(n)
	for temp.Sign() > 0 {
		// Find the position of the lowest set bit
		trailingZeros := temp.TrailingZeroBits()
		indices = append(indices, int(trailingZeros))
		// Clear the lowest set bit using Brian Kernighan's algorithm
		temp.And(temp, new(big.Int).Sub(temp, big.NewInt(1)))
	}
	return indices
}

// Optimized for dense bit patterns (many set bits)
func getSetBitIndicesDense(n *big.Int, bitLen int) []int {
	var indices []int
	// Pre-allocate with estimated capacity to reduce allocations
	indices = make([]int, 0, bitLen/4) // Conservative estimate
	for i := range bitLen {
		if n.Bit(i) == 1 {
			indices = append(indices, i)
		}
	}
	return indices
}

func estimateBitDensity(n *big.Int, bitLen int) float64 {
	if bitLen <= 100 {
		// For small numbers, count all bits
		setBits := 0
		for i := range bitLen {
			if n.Bit(i) == 1 {
				setBits++
			}
		}
		return float64(setBits) / float64(bitLen)
	}
	// Sample approach for large numbers
	sampleSize := min(100, bitLen/10)
	setBits := 0
	step := bitLen / sampleSize
	for i := range sampleSize {
		pos := i * step
		if pos < bitLen && n.Bit(pos) == 1 {
			setBits++
		}
	}
	return float64(setBits) / float64(sampleSize)
}

func shouldUseSparseAlgorithm(density float64, bitLen int) bool {
	// Use sparse algorithm if:
	// 1. Density is low (< 25%)
	// 2. For very large numbers with medium density (< 50%), sparse is still better
	// 3. Adaptive threshold based on bit length
	threshold := 0.25
	if bitLen > 1000 {
		threshold = 0.4
	}
	if bitLen > 10000 {
		threshold = 0.5
	}
	return density < threshold
}
