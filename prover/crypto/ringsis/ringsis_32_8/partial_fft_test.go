// Code generated by bavard DO NOT EDIT

package ringsis_32_8

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/assert"
	"github.com/consensys/linea-monorepo/prover/maths/fft"
)

func TestPartialFFT(t *testing.T) {

	var (
		domain   = fft.NewDomain(32).WithCoset()
		twiddles = PrecomputeTwiddlesCoset(domain.Generator, domain.FrMultiplicativeGen)
	)

	for mask := 0; mask < 2; mask++ {

		var (
			a = vec123456()
			b = vec123456()
		)

		zeroizeWithMask(a, mask)
		zeroizeWithMask(b, mask)

		domain.FFT(a, fft.DIF, fft.OnCoset())
		partialFFT[mask](b, twiddles)
		assert.Equal(t, a, b)
	}

}

func vec123456() []field.Element {
	vec := make([]field.Element, 32)
	for i := range vec {
		vec[i].SetInt64(int64(i))
	}
	return vec
}

func zeroizeWithMask(v []field.Element, mask int) {
	for i := 0; i < 1; i++ {
		if (mask>>i)&1 == 1 {
			continue
		}

		for j := 0; j < 32; j++ {
			v[32*i+j].SetZero()
		}
	}
}
