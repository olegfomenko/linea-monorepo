package fft

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"math/bits"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Decimation is used in the FFT call to select decimation in time or in frequency
type Decimation uint8

const (
	DIT Decimation = iota
	DIF
)

// parallelize threshold for a single butterfly op, if the fft stage is not parallelized already
const butterflyThreshold = 16

// FFT computes (recursively) the discrete Fourier transform of a and stores the result in a
// if decimation == DIT (decimation in time), the input must be in bit-reversed order
// if decimation == DIF (decimation in frequency), the output will be in bit-reversed order
// if coset if set, the FFT(a) returns the evaluation of a on a coset.
func (domain *Domain) FFT(a []field.Element, decimation Decimation, opts ...Option) {

	opt := fftOptions(opts...)

	// find the stage where we should stop spawning go routines in our recursive calls
	// (ie when we have as many go routines running as we have available CPUs)
	maxSplits := bits.TrailingZeros64(ecc.NextPowerOfTwo(uint64(opt.nbTasks)))
	if opt.nbTasks == 1 {
		maxSplits = -1
	}

	// if coset != 0, scale by coset table
	if opt.coset {
		scale := func(cosetTable []field.Element) {
			for i := 0; i < len(a); i++ {
				a[i].Mul(&a[i], &cosetTable[i])
			}
		}
		if decimation == DIT {
			scale(domain.CosetTableReversed)

		} else {
			scale(domain.CosetTable)
		}
	}

	switch decimation {
	case DIF:
		difFFTIterable(a, domain.Twiddles, maxSplits, opt.nbTasks)
	case DIT:
		ditFFTIterable(a, domain.Twiddles, maxSplits, opt.nbTasks)
	default:
		panic("not implemented")
	}
}

// FFTInverse computes (recursively) the inverse discrete Fourier transform of a and stores the result in a
// if decimation == DIT (decimation in time), the input must be in bit-reversed order
// if decimation == DIF (decimation in frequency), the output will be in bit-reversed order
// coset sets the shift of the fft (0 = no shift, standard fft)
// len(a) must be a power of 2, and w must be a len(a)th root of unity in field F.
func (domain *Domain) FFTInverse(a []field.Element, decimation Decimation, opts ...Option) {

	opt := fftOptions(opts...)

	// find the stage where we should stop spawning go routines in our recursive calls
	// (ie when we have as many go routines running as we have available CPUs)
	maxSplits := bits.TrailingZeros64(ecc.NextPowerOfTwo(uint64(opt.nbTasks)))
	if opt.nbTasks == 1 {
		maxSplits = -1
	}

	switch decimation {
	case DIF:
		difFFTIterable(a, domain.TwiddlesInv, maxSplits, opt.nbTasks)
	case DIT:
		ditFFTIterable(a, domain.TwiddlesInv, maxSplits, opt.nbTasks)
	default:
		panic("not implemented")
	}

	// scale by CardinalityInv
	if !opt.coset {
		for i := 0; i < len(a); i++ {
			a[i].Mul(&a[i], &domain.CardinalityInv)
		}
		return
	}

	scale := func(cosetTable []field.Element) {
		for i := 0; i < len(a); i++ {
			a[i].Mul(&a[i], &cosetTable[i]).
				Mul(&a[i], &domain.CardinalityInv)
		}
	}
	if decimation == DIT {
		scale(domain.CosetTableInv)
		return
	}

	// decimation == DIF
	scale(domain.CosetTableInvReversed)

}

func difFFTIterable(a []field.Element, twiddles [][]field.Element, maxSplits int, nbTasks int) {
	n := len(a)

	stage := 0

	var m int
	iterations := utils.Log2Floor(n)
	for i := 0; i < iterations; i++ {
		if n == 1 {
			return
		} else if n == 256 {
			parallel.Execute(1<<stage, func(start, end int) {
				for j := start; j < end; j++ {
					kerDIFNP_256(a[j*n:(j+1)*n], twiddles, stage)
				}
			})

			return
		}

		m = n >> 1

		parallelButterfly := (m > butterflyThreshold) && (stage < maxSplits)
		parallel.Execute(1<<stage, func(start, end int) {
			for j := start; j < end; j++ {
				b := a[j*n : (j+1)*n]

				if !parallelButterfly {
					innerDIFWithTwiddles(b, twiddles[stage], 0, m, m)
					continue
				}

				parallel.Execute(m, func(start, end int) {
					innerDIFWithTwiddles(b, twiddles[stage], start, end, m)
				})
			}
		})

		stage++
		n = m
	}
}

func difFFT(a []field.Element, twiddles [][]field.Element, stage int, maxSplits int, chDone chan struct{}, nbTasks int) {
	if chDone != nil {
		defer close(chDone)
	}

	n := len(a)
	if n == 1 {
		return
	} else if n == 256 {
		kerDIFNP_256(a, twiddles, stage)
		return
	}

	m := n >> 1

	parallelButterfly := (m > butterflyThreshold) && (stage < maxSplits)

	// i == 0
	if parallelButterfly {
		parallel.Execute(m, func(start, end int) {
			innerDIFWithTwiddles(a, twiddles[stage], start, end, m)
		}, nbTasks/(1<<(stage)))
	} else {
		innerDIFWithTwiddles(a, twiddles[stage], 0, m, m)
	}

	if m == 1 {
		return
	}

	nextStage := stage + 1
	if stage < maxSplits {
		chDone := make(chan struct{}, 1)
		go difFFT(a[m:n], twiddles, nextStage, maxSplits, chDone, nbTasks)
		difFFT(a[0:m], twiddles, nextStage, maxSplits, nil, nbTasks)
		<-chDone
	} else {
		difFFT(a[0:m], twiddles, nextStage, maxSplits, nil, nbTasks)
		difFFT(a[m:n], twiddles, nextStage, maxSplits, nil, nbTasks)
	}
}

func ditFFT(a []field.Element, twiddles [][]field.Element, stage int, maxSplits int, chDone chan struct{}, nbTasks int) {
	if chDone != nil {
		defer close(chDone)
	}

	n := len(a)
	if n == 1 {
		return
	} else if n == 256 {
		kerDITNP_256(a, twiddles, stage)
		return
	}

	m := n >> 1

	parallelButterfly := (m > butterflyThreshold) && (stage < maxSplits)

	nextStage := stage + 1

	if stage < maxSplits {
		// that's the only time we fire go routines
		chDone := make(chan struct{}, 1)
		go ditFFT(a[m:n], twiddles, nextStage, maxSplits, chDone, nbTasks)
		ditFFT(a[0:m], twiddles, nextStage, maxSplits, nil, nbTasks)
		<-chDone
	} else {
		ditFFT(a[0:m], twiddles, nextStage, maxSplits, nil, nbTasks)
		ditFFT(a[m:n], twiddles, nextStage, maxSplits, nil, nbTasks)
	}

	if parallelButterfly {
		parallel.Execute(m, func(start, end int) {
			innerDITWithTwiddles(a, twiddles[stage], start, end, m)
		}, nbTasks/(1<<(stage)))
	} else {
		innerDITWithTwiddles(a, twiddles[stage], 0, m, m)
	}
}

func ditFFTIterable(a []field.Element, twiddles [][]field.Element, maxSplits int, nbTasks int) {
	if len(a) == 1 {
		return
	}

	n := 2
	iterations := utils.Log2Floor(len(a))
	if len(a) >= 256 {
		n = 256
		iterations -= 8

		parallel.ExecuteChunky(1<<(iterations), func(start, end int) {
			for j := start; j < end; j++ {
				kerDITNP_256(a[j*n:(j+1)*n], twiddles, iterations)
			}
		})

		n <<= 1
	}

	for i := iterations - 1; i >= 0; i-- {
		m := n >> 1

		if len(a) < 256 {
			for j := 0; j < 1<<i; j++ {
				innerDITWithTwiddlesAndOffset(a, twiddles[i], 0, m, m, j*n)
			}
		} else {
			// Number of subchunks in chunk.
			// TODO: determine how to adjust this values using maxSplits and nbTasks.
			subchunksInChunk := 2

			// The number of chunks for the current FFT level == len(a)/n
			chunksNum := i + 1 // == len(a)/n
			subChunksNum := chunksNum * subchunksInChunk

			parallel.Execute(subChunksNum, func(start, end int) {
				// Here j is an index of a subchunk.
				for j := start; j < end; j++ {
					// Identify the main chunk for this subchunk.
					chunk := j / subchunksInChunk
					// Calculate the position of this subchunk within its chunk.
					subchunk := j % subchunksInChunk
					// Determine the starting offset for the chunk in array 'a'.
					offset := chunk * n
					// The size of subchunk is a half of each chunk divided by subchunk count.
					subchunkSize := n / (2 * subchunksInChunk)

					innerDITWithTwiddlesAndOffset(a, twiddles[i], subchunk*subchunkSize,
						(subchunk+1)*subchunkSize, m, offset)
				}
			})
		}

		// Double the chunk size for the next FFT level.
		n <<= 1
	}
}

func innerDIFWithTwiddles(a []field.Element, twiddles []field.Element, start, end, m int) {
	if start == 0 {
		field.Butterfly(&a[0], &a[m])
		start++
	}
	for i := start; i < end; i++ {
		field.Butterfly(&a[i], &a[i+m])
		a[i+m].Mul(&a[i+m], &twiddles[i])
	}
}

func innerDITWithTwiddles(a []field.Element, twiddles []field.Element, start, end, m int) {
	if start == 0 {
		field.Butterfly(&a[0], &a[m])
		start++
	}
	for i := start; i < end; i++ {
		a[i+m].Mul(&a[i+m], &twiddles[i])
		field.Butterfly(&a[i], &a[i+m])
	}
}

func innerDITWithTwiddlesAndOffset(a []field.Element, twiddles []field.Element, start, end, m, offset int) {
	if start == 0 {
		field.Butterfly(&a[offset], &a[offset+m])
		start++
	}
	for i := start; i < end; i++ {
		a[offset+i+m].Mul(&a[offset+i+m], &twiddles[i])
		field.Butterfly(&a[offset+i], &a[offset+i+m])
	}
}

func kerDIFNP_256(a []field.Element, twiddles [][]field.Element, stage int) {
	// code unrolled & generated by internal/generator/fft/template/fft.go.tmpl

	innerDIFWithTwiddles(a[:256], twiddles[stage+0], 0, 128, 128)
	for offset := 0; offset < 256; offset += 128 {
		innerDIFWithTwiddles(a[offset:offset+128], twiddles[stage+1], 0, 64, 64)
	}
	for offset := 0; offset < 256; offset += 64 {
		innerDIFWithTwiddles(a[offset:offset+64], twiddles[stage+2], 0, 32, 32)
	}
	for offset := 0; offset < 256; offset += 32 {
		innerDIFWithTwiddles(a[offset:offset+32], twiddles[stage+3], 0, 16, 16)
	}
	for offset := 0; offset < 256; offset += 16 {
		innerDIFWithTwiddles(a[offset:offset+16], twiddles[stage+4], 0, 8, 8)
	}
	for offset := 0; offset < 256; offset += 8 {
		innerDIFWithTwiddles(a[offset:offset+8], twiddles[stage+5], 0, 4, 4)
	}
	for offset := 0; offset < 256; offset += 4 {
		innerDIFWithTwiddles(a[offset:offset+4], twiddles[stage+6], 0, 2, 2)
	}
	for offset := 0; offset < 256; offset += 2 {
		field.Butterfly(&a[offset], &a[offset+1])
	}
}

func kerDITNP_256(a []field.Element, twiddles [][]field.Element, stage int) {
	// code unrolled & generated by internal/generator/fft/template/fft.go.tmpl

	for offset := 0; offset < 256; offset += 2 {
		field.Butterfly(&a[offset], &a[offset+1])
	}
	for offset := 0; offset < 256; offset += 4 {
		innerDITWithTwiddles(a[offset:offset+4], twiddles[stage+6], 0, 2, 2)
	}
	for offset := 0; offset < 256; offset += 8 {
		innerDITWithTwiddles(a[offset:offset+8], twiddles[stage+5], 0, 4, 4)
	}
	for offset := 0; offset < 256; offset += 16 {
		innerDITWithTwiddles(a[offset:offset+16], twiddles[stage+4], 0, 8, 8)
	}
	for offset := 0; offset < 256; offset += 32 {
		innerDITWithTwiddles(a[offset:offset+32], twiddles[stage+3], 0, 16, 16)
	}
	for offset := 0; offset < 256; offset += 64 {
		innerDITWithTwiddles(a[offset:offset+64], twiddles[stage+2], 0, 32, 32)
	}
	for offset := 0; offset < 256; offset += 128 {
		innerDITWithTwiddles(a[offset:offset+128], twiddles[stage+1], 0, 64, 64)
	}
	innerDITWithTwiddles(a[:256], twiddles[stage+0], 0, 128, 128)
}
