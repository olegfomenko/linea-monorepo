package fastpoly

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/require"
)

func TestInterpolation(t *testing.T) {
	n := 4
	randPoly := vector.ForTest(1, 2, 3, 4)
	x := field.NewElement(51)
	expectedY := poly.EvalUnivariate(randPoly, x)
	domain := fft.NewDomain(n).WithCoset()

	/*
		Test without coset
	*/
	onRoots := vector.DeepCopy(randPoly)
	domain.FFT(onRoots, fft.DIF)

	fft.BitReverse(onRoots)
	yOnRoots := Interpolate(onRoots, x)
	require.Equal(t, expectedY.String(), yOnRoots.String())

	/*
		Test with coset
	*/
	onCoset := vector.DeepCopy(randPoly)
	domain.FFT(onCoset, fft.DIF, fft.OnCoset())
	fft.BitReverse(onCoset)
	yOnCoset := Interpolate(onCoset, x, true)
	require.Equal(t, expectedY.String(), yOnCoset.String())

}

// BenchmarkInterpolation-12    	     250	   4648674 ns/op
// BenchmarkInterpolation-12    	     253	   4764962 ns/op
func BenchmarkInterpolation(b *testing.B) {
	rand := rand.New(rand.NewSource(786868))

	for i := 8; i < 22; i++ {
		n := 1 << i

		b.Run(fmt.Sprintf("Vanilla without coset %d", i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				randPoly := vector.Rand(n)
				x := field.PseudoRand(rand)
				expected := poly.EvalUnivariate(randPoly, x)
				domain := fft.NewDomain(n).WithCoset()

				onRoots := vector.DeepCopy(randPoly)
				domain.FFT(onRoots, fft.DIF)

				fft.BitReverse(onRoots)

				b.StartTimer()
				yOnRoots := Interpolate(onRoots, x)
				b.StopTimer()

				require.Equal(b, expected.String(), yOnRoots.String())

				/*
					Test with coset
				*/
				//onCoset := vector.DeepCopy(randPoly)
				//domain.FFT(onCoset, fft.DIF, fft.OnCoset())
				//fft.BitReverse(onCoset)
				//b.StartTimer()
				//yOnCoset := Interpolate(onCoset, x, true)
				//b.StopTimer()
				//require.Equal(b, expected.String(), yOnCoset.String())
			}
		})

		b.Run(fmt.Sprintf("Vanilla on coset %d", i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				randPoly := vector.Rand(n)
				x := field.PseudoRand(rand)
				expected := poly.EvalUnivariate(randPoly, x)
				domain := fft.NewDomain(n).WithCoset()

				onCoset := vector.DeepCopy(randPoly)
				domain.FFT(onCoset, fft.DIF, fft.OnCoset())
				fft.BitReverse(onCoset)
				b.StartTimer()
				yOnCoset := Interpolate(onCoset, x, true)
				b.StopTimer()
				require.Equal(b, expected.String(), yOnCoset.String())
			}
		})

		b.Run(fmt.Sprintf("New without coset %d", i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				randPoly := vector.Rand(n)
				x := field.PseudoRand(rand)
				expected := poly.EvalUnivariate(randPoly, x)
				domain := fft.NewDomain(n).WithCoset()

				onRoots := vector.DeepCopy(randPoly)
				domain.FFT(onRoots, fft.DIF)

				fft.BitReverse(onRoots)

				b.StartTimer()
				yOnRoots := NewInterpolate(onRoots, x)
				b.StopTimer()

				require.Equal(b, expected.String(), yOnRoots.String())

				/*
					Test with coset
				*/
				//onCoset := vector.DeepCopy(randPoly)
				//domain.FFT(onCoset, fft.DIF, fft.OnCoset())
				//fft.BitReverse(onCoset)
				//b.StartTimer()
				//yOnCoset := NewInterpolate(onCoset, x, true)
				//b.StopTimer()
				//require.Equal(b, expected.String(), yOnCoset.String())

			}
		})

		b.Run(fmt.Sprintf("New on coset %d", i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				randPoly := vector.Rand(n)
				x := field.PseudoRand(rand)
				expected := poly.EvalUnivariate(randPoly, x)
				domain := fft.NewDomain(n).WithCoset()

				onCoset := vector.DeepCopy(randPoly)
				domain.FFT(onCoset, fft.DIF, fft.OnCoset())
				fft.BitReverse(onCoset)
				b.StartTimer()
				yOnCoset := NewInterpolate(onCoset, x, true)
				b.StopTimer()
				require.Equal(b, expected.String(), yOnCoset.String())
			}
		})
	}

}

func TestBatchInterpolation(t *testing.T) {
	n := 4
	randPoly := vector.ForTest(1, 2, 3, 4)
	randPoly2 := vector.ForTest(5, 6, 7, 8)
	x := field.NewElement(51)

	expectedY := poly.EvalUnivariate(randPoly, x)
	expectedY2 := poly.EvalUnivariate(randPoly2, x)
	domain := fft.NewDomain(n).WithCoset()

	/*
		Test without coset
	*/
	onRoots := vector.DeepCopy(randPoly)
	onRoots2 := vector.DeepCopy(randPoly2)
	polys := make([][]field.Element, 2)
	polys[0] = onRoots
	polys[1] = onRoots2

	domain.FFT(polys[0], fft.DIF)
	domain.FFT(polys[1], fft.DIF)
	fft.BitReverse(polys[0])
	fft.BitReverse(polys[1])

	yOnRoots := BatchInterpolate(polys, x)
	require.Equal(t, expectedY.String(), yOnRoots[0].String())
	require.Equal(t, expectedY2.String(), yOnRoots[1].String())

	/*
		Test with coset
	*/
	onCoset := vector.DeepCopy(randPoly)
	onCoset2 := vector.DeepCopy(randPoly2)
	onCosets := make([][]field.Element, 2)
	onCosets[0] = onCoset
	onCosets[1] = onCoset2

	domain.FFT(onCosets[0], fft.DIF, fft.OnCoset())
	domain.FFT(onCosets[1], fft.DIF, fft.OnCoset())
	fft.BitReverse(onCosets[0])
	fft.BitReverse(onCosets[1])

	yOnCosets := BatchInterpolate(onCosets, x, true)
	require.Equal(t, expectedY.String(), yOnCosets[0].String())
	require.Equal(t, expectedY2.String(), yOnCosets[1].String())

}

// edge-case : x is a root of unity of the domain. In this case, we can just return
// the associated value for poly
func TestBatchInterpolationRootOfUnity(t *testing.T) {
	n := 4
	randPoly := vector.ForTest(1, 2, 3, 4)
	randPoly2 := vector.ForTest(5, 6, 7, 8)

	// define x as a root of unity
	x := field.One()

	expectedY := poly.EvalUnivariate(randPoly, x)
	expectedY2 := poly.EvalUnivariate(randPoly2, x)
	domain := fft.NewDomain(n).WithCoset()

	/*
		Test without coset
	*/
	onRoots := vector.DeepCopy(randPoly)
	onRoots2 := vector.DeepCopy(randPoly2)
	polys := make([][]field.Element, 2)
	polys[0] = onRoots
	polys[1] = onRoots2

	domain.FFT(polys[0], fft.DIF)
	domain.FFT(polys[1], fft.DIF)
	fft.BitReverse(polys[0])
	fft.BitReverse(polys[1])

	yOnRoots := BatchInterpolate(polys, x)
	require.Equal(t, expectedY.String(), yOnRoots[0].String())
	require.Equal(t, expectedY2.String(), yOnRoots[1].String())

	/*
		Test with coset
	*/
	onCoset := vector.DeepCopy(randPoly)
	onCoset2 := vector.DeepCopy(randPoly2)
	onCosets := make([][]field.Element, 2)
	onCosets[0] = onCoset
	onCosets[1] = onCoset2

	domain.FFT(onCosets[0], fft.DIF, fft.OnCoset())
	domain.FFT(onCosets[1], fft.DIF, fft.OnCoset())
	fft.BitReverse(onCosets[0])
	fft.BitReverse(onCosets[1])

	yOnCosets := BatchInterpolate(onCosets, x, true)
	require.Equal(t, expectedY.String(), yOnCosets[0].String())
	require.Equal(t, expectedY2.String(), yOnCosets[1].String())

}
