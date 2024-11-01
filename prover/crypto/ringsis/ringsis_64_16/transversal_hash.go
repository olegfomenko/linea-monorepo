// Code generated by bavard DO NOT EDIT

package ringsis_64_16

import (
	"runtime"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	ppool "github.com/consensys/linea-monorepo/prover/utils/parallel/pool"
)

func TransversalHash(
	// the Ag for ring-sis
	ag [][]field.Element,
	// A non-transposed list of columns
	// All of the same length
	pols []smartvectors.SmartVector,
	// The precomputed twiddle cosets for the forward FFT
	twiddleCosets []field.Element,
	// The domain for the final inverse-FFT
	domain *fft.Domain,
) []field.Element {

	var (
		// Each field element is encoded in 16 limbs but the degree is 64. So, each
		// polynomial multiplication "hashes" 4 field elements at once. This is
		// important to know for parallelization.
		resultSize = pols[0].Len() * 64

		// To optimize memory usage, we limit ourself to hash only 16 columns per
		// iteration.
		numColumnPerJob int = 16

		// In theory, it should be a div ceil. But in practice we only process power's
		// of two number of columns. If that's not the case, then the function will panic
		// but we can always change that if this is needed. The rational for the current
		// design is simplicity.
		numJobs = utils.DivExact(pols[0].Len(), numColumnPerJob) // we make blocks of 16 columns

		// Main result of the hashing
		mainResults = make([]field.Element, resultSize)
		// When we encounter a const row, it will have the same additive contribution
		// to the result on every column. So we compute the contribution only once and
		// accumulate it with the other "constant column contributions". And it is only
		// performed by the first thread.
		constResults = make([]field.Element, 64)
	)

	ppool.ExecutePoolChunky(numJobs, func(i int) {

		var (
			localResult = make([]field.Element, numColumnPerJob*64)
			limbs       = make([]field.Element, 64)

			// Each segment is processed by packet of `numFieldPerPoly=4` rows
			startFromCol = i * numColumnPerJob
			stopAtCol    = (i + 1) * numColumnPerJob
		)

		for row := 0; row < len(pols); row += 4 {

			var (
				chunksFull = make([][]field.Element, 4)
				mask       = 0
			)

			for j := 0; j < 4; j++ {
				if row+j >= len(pols) {
					continue
				}

				pReg, pIsReg := pols[row+j].(*smartvectors.Regular)
				if pIsReg {
					chunksFull[j] = (*pReg)[startFromCol:stopAtCol]
					mask |= (1 << j)
					continue
				}

				pPool, pIsPool := pols[row+j].(*smartvectors.Pooled)
				if pIsPool {
					chunksFull[j] = pPool.Regular[startFromCol:stopAtCol]
					mask |= (1 << j)
					continue
				}
			}

			if mask > 0 {
				for col := 0; col < (stopAtCol - startFromCol); col++ {
					colChunk := [4]field.Element{}
					for j := 0; j < 4; j++ {
						if chunksFull[j] != nil {
							colChunk[j] = chunksFull[j][col]
						}
					}

					limbDecompose(limbs, colChunk[:])
					partialFFT[mask](limbs, twiddleCosets)
					mulModAcc(localResult[col*64:(col+1)*64], limbs, ag[row/4])
				}
			}

			if i == 0 {

				var (
					cMask      = ((1 << 4) - 1) ^ mask
					chunkConst = make([]field.Element, 4)
				)

				if cMask > 0 {
					for j := 0; j < 4; j++ {
						if row+j >= len(pols) {
							continue
						}

						if (cMask>>j)&1 == 1 {
							chunkConst[j] = pols[row+j].(*smartvectors.Constant).Get(0)
						}
					}

					limbDecompose(limbs, chunkConst)
					partialFFT[cMask](limbs, twiddleCosets)
					mulModAcc(constResults, limbs, ag[row/4])
				}
			}
		}

		// copy the segment into the main result at the end
		copy(mainResults[startFromCol*64:stopAtCol*64], localResult)
	})

	// Now, we need to reconciliate the results of the buffer with
	// the result for each thread
	parallel.Execute(pols[0].Len(), func(start, stop int) {
		for col := start; col < stop; col++ {
			// Accumulate the const
			vector.Add(mainResults[col*64:(col+1)*64], mainResults[col*64:(col+1)*64], constResults)
			// And run the reverse FFT
			domain.FFTInverse(mainResults[col*64:(col+1)*64], fft.DIT, fft.OnCoset(), fft.WithNbTasks(1))
		}
	})

	return mainResults
}

var _zeroes []field.Element = make([]field.Element, 64)

// zeroize fills `buf` with zeroes.
func zeroize(buf []field.Element) {
	copy(buf, _zeroes)
}

// mulModAdd increments each entry `i` of `res` as `res[i] = a[i] * b[i]`. The
// input vectors are trusted to all have the same length.
func mulModAcc(res, a, b []field.Element) {
	var tmp field.Element
	for i := range res {
		tmp.Mul(&a[i], &b[i])
		res[i].Add(&res[i], &tmp)
	}
}

func limbDecompose(result []field.Element, x []field.Element) {

	zeroize(result)
	var bytesBuffer = [32]byte{}

	bytesBuffer = x[0].Bytes()

	result[15][0] = uint64(bytesBuffer[1]) | (uint64(bytesBuffer[0]) << 8)
	result[14][0] = uint64(bytesBuffer[3]) | (uint64(bytesBuffer[2]) << 8)
	result[13][0] = uint64(bytesBuffer[5]) | (uint64(bytesBuffer[4]) << 8)
	result[12][0] = uint64(bytesBuffer[7]) | (uint64(bytesBuffer[6]) << 8)
	result[11][0] = uint64(bytesBuffer[9]) | (uint64(bytesBuffer[8]) << 8)
	result[10][0] = uint64(bytesBuffer[11]) | (uint64(bytesBuffer[10]) << 8)
	result[9][0] = uint64(bytesBuffer[13]) | (uint64(bytesBuffer[12]) << 8)
	result[8][0] = uint64(bytesBuffer[15]) | (uint64(bytesBuffer[14]) << 8)
	result[7][0] = uint64(bytesBuffer[17]) | (uint64(bytesBuffer[16]) << 8)
	result[6][0] = uint64(bytesBuffer[19]) | (uint64(bytesBuffer[18]) << 8)
	result[5][0] = uint64(bytesBuffer[21]) | (uint64(bytesBuffer[20]) << 8)
	result[4][0] = uint64(bytesBuffer[23]) | (uint64(bytesBuffer[22]) << 8)
	result[3][0] = uint64(bytesBuffer[25]) | (uint64(bytesBuffer[24]) << 8)
	result[2][0] = uint64(bytesBuffer[27]) | (uint64(bytesBuffer[26]) << 8)
	result[1][0] = uint64(bytesBuffer[29]) | (uint64(bytesBuffer[28]) << 8)
	result[0][0] = uint64(bytesBuffer[31]) | (uint64(bytesBuffer[30]) << 8)

	bytesBuffer = x[1].Bytes()

	result[31][0] = uint64(bytesBuffer[1]) | (uint64(bytesBuffer[0]) << 8)
	result[30][0] = uint64(bytesBuffer[3]) | (uint64(bytesBuffer[2]) << 8)
	result[29][0] = uint64(bytesBuffer[5]) | (uint64(bytesBuffer[4]) << 8)
	result[28][0] = uint64(bytesBuffer[7]) | (uint64(bytesBuffer[6]) << 8)
	result[27][0] = uint64(bytesBuffer[9]) | (uint64(bytesBuffer[8]) << 8)
	result[26][0] = uint64(bytesBuffer[11]) | (uint64(bytesBuffer[10]) << 8)
	result[25][0] = uint64(bytesBuffer[13]) | (uint64(bytesBuffer[12]) << 8)
	result[24][0] = uint64(bytesBuffer[15]) | (uint64(bytesBuffer[14]) << 8)
	result[23][0] = uint64(bytesBuffer[17]) | (uint64(bytesBuffer[16]) << 8)
	result[22][0] = uint64(bytesBuffer[19]) | (uint64(bytesBuffer[18]) << 8)
	result[21][0] = uint64(bytesBuffer[21]) | (uint64(bytesBuffer[20]) << 8)
	result[20][0] = uint64(bytesBuffer[23]) | (uint64(bytesBuffer[22]) << 8)
	result[19][0] = uint64(bytesBuffer[25]) | (uint64(bytesBuffer[24]) << 8)
	result[18][0] = uint64(bytesBuffer[27]) | (uint64(bytesBuffer[26]) << 8)
	result[17][0] = uint64(bytesBuffer[29]) | (uint64(bytesBuffer[28]) << 8)
	result[16][0] = uint64(bytesBuffer[31]) | (uint64(bytesBuffer[30]) << 8)

	bytesBuffer = x[2].Bytes()

	result[47][0] = uint64(bytesBuffer[1]) | (uint64(bytesBuffer[0]) << 8)
	result[46][0] = uint64(bytesBuffer[3]) | (uint64(bytesBuffer[2]) << 8)
	result[45][0] = uint64(bytesBuffer[5]) | (uint64(bytesBuffer[4]) << 8)
	result[44][0] = uint64(bytesBuffer[7]) | (uint64(bytesBuffer[6]) << 8)
	result[43][0] = uint64(bytesBuffer[9]) | (uint64(bytesBuffer[8]) << 8)
	result[42][0] = uint64(bytesBuffer[11]) | (uint64(bytesBuffer[10]) << 8)
	result[41][0] = uint64(bytesBuffer[13]) | (uint64(bytesBuffer[12]) << 8)
	result[40][0] = uint64(bytesBuffer[15]) | (uint64(bytesBuffer[14]) << 8)
	result[39][0] = uint64(bytesBuffer[17]) | (uint64(bytesBuffer[16]) << 8)
	result[38][0] = uint64(bytesBuffer[19]) | (uint64(bytesBuffer[18]) << 8)
	result[37][0] = uint64(bytesBuffer[21]) | (uint64(bytesBuffer[20]) << 8)
	result[36][0] = uint64(bytesBuffer[23]) | (uint64(bytesBuffer[22]) << 8)
	result[35][0] = uint64(bytesBuffer[25]) | (uint64(bytesBuffer[24]) << 8)
	result[34][0] = uint64(bytesBuffer[27]) | (uint64(bytesBuffer[26]) << 8)
	result[33][0] = uint64(bytesBuffer[29]) | (uint64(bytesBuffer[28]) << 8)
	result[32][0] = uint64(bytesBuffer[31]) | (uint64(bytesBuffer[30]) << 8)

	bytesBuffer = x[3].Bytes()

	result[63][0] = uint64(bytesBuffer[1]) | (uint64(bytesBuffer[0]) << 8)
	result[62][0] = uint64(bytesBuffer[3]) | (uint64(bytesBuffer[2]) << 8)
	result[61][0] = uint64(bytesBuffer[5]) | (uint64(bytesBuffer[4]) << 8)
	result[60][0] = uint64(bytesBuffer[7]) | (uint64(bytesBuffer[6]) << 8)
	result[59][0] = uint64(bytesBuffer[9]) | (uint64(bytesBuffer[8]) << 8)
	result[58][0] = uint64(bytesBuffer[11]) | (uint64(bytesBuffer[10]) << 8)
	result[57][0] = uint64(bytesBuffer[13]) | (uint64(bytesBuffer[12]) << 8)
	result[56][0] = uint64(bytesBuffer[15]) | (uint64(bytesBuffer[14]) << 8)
	result[55][0] = uint64(bytesBuffer[17]) | (uint64(bytesBuffer[16]) << 8)
	result[54][0] = uint64(bytesBuffer[19]) | (uint64(bytesBuffer[18]) << 8)
	result[53][0] = uint64(bytesBuffer[21]) | (uint64(bytesBuffer[20]) << 8)
	result[52][0] = uint64(bytesBuffer[23]) | (uint64(bytesBuffer[22]) << 8)
	result[51][0] = uint64(bytesBuffer[25]) | (uint64(bytesBuffer[24]) << 8)
	result[50][0] = uint64(bytesBuffer[27]) | (uint64(bytesBuffer[26]) << 8)
	result[49][0] = uint64(bytesBuffer[29]) | (uint64(bytesBuffer[28]) << 8)
	result[48][0] = uint64(bytesBuffer[31]) | (uint64(bytesBuffer[30]) << 8)
}
