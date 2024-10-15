// Copyright 2023 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sis

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash"
	"math/big"
	"math/bits"

	"github.com/bits-and-blooms/bitset"
	"golang.org/x/crypto/blake2b"

	"github.com/consensys/linea-monorepo/prover/maths/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// Ring-SIS instance
type RSis struct {

	// buffer storing the data to hash
	buffer bytes.Buffer

	// Vectors in ℤ_{p}/Xⁿ+1
	// A[i] is the i-th polynomial.
	// Ag the evaluation form of the polynomials in A on the coset √(g) * <g>
	A  [][]field.Element
	Ag [][]field.Element

	// LogTwoBound (Infinity norm) of the vector to hash. It means that each component in m
	// is < 2^B, where m is the vector to hash (the hash being A*m).
	// cf https://hackmd.io/7OODKWQZRRW9RxM5BaXtIw , B >= 3.
	LogTwoBound int

	// domain for the polynomial multiplication
	Domain        *fft.Domain
	twiddleCosets []field.Element // see FFT64 and precomputeTwiddlesCoset

	// d, the degree of X^{d}+1
	Degree int

	// in bytes, represents the maximum number of bytes the .Write(...) will handle;
	// ( maximum number of bytes to sum )
	capacity            int
	maxNbElementsToHash int

	// allocate memory once per instance (used in Sum())
	bufM, bufRes []field.Element
	bufMValues   *bitset.BitSet
}

// NewRSis creates an instance of RSis.
// seed: seed for the randomness for generating A.
// logTwoDegree: if d := logTwoDegree, the ring will be ℤ_{p}[X]/Xᵈ-1, where X^{2ᵈ} is the 2ᵈ⁺¹-th cyclotomic polynomial
// logTwoBound: the bound of the vector to hash (using the infinity norm).
// maxNbElementsToHash: maximum number of field elements the instance handles
// used to derived n, the number of polynomials in A, and max size of instance's internal buffer.
func NewRSis(seed int64, logTwoDegree, logTwoBound, maxNbElementsToHash int) (*RSis, error) {

	if logTwoBound > 64 {
		return nil, errors.New("logTwoBound too large")
	}
	if bits.UintSize == 32 {
		return nil, errors.New("unsupported architecture; need 64bit target")
	}

	degree := 1 << logTwoDegree
	capacity := maxNbElementsToHash * field.Bytes

	// n: number of polynomials in A
	// len(m) == degree * n
	// with each element in m being logTwoBounds bits from the instance buffer.
	// that is, to fill m, we need [degree * n * logTwoBound] bits of data
	// capacity == [degree * n * logTwoBound] / 8
	// n == (capacity*8)/(degree*logTwoBound)

	// First n <- #limbs to represent a single field element
	n := (field.Bytes * 8) / logTwoBound
	if n*logTwoBound < field.Bytes*8 {
		n++
	}

	// Then multiply by the number of field elements
	n *= maxNbElementsToHash

	// And divide (+ ceil) to get the number of polynomials
	if n%degree == 0 {
		n /= degree
	} else {
		n /= degree // number of polynomials
		n++
	}

	// domains (shift is √{gen} )
	var shift field.Element
	e := int64(1 << (47 - (logTwoDegree + 1)))
	shift.Exp(field.RootOfUnity, big.NewInt(e))

	r := &RSis{
		LogTwoBound:         logTwoBound,
		capacity:            capacity,
		Degree:              degree,
		Domain:              fft.NewDomain(degree).WithShift(shift),
		A:                   make([][]field.Element, n),
		Ag:                  make([][]field.Element, n),
		bufM:                make([]field.Element, degree*n),
		bufRes:              make([]field.Element, degree),
		bufMValues:          bitset.New(uint(n)),
		maxNbElementsToHash: maxNbElementsToHash,
	}
	if r.LogTwoBound == 8 && r.Degree == 64 {
		// TODO @gbotrel fixme, that's dirty.
		r.twiddleCosets = PrecomputeTwiddlesCoset(r.Domain.Generator, r.Domain.FrMultiplicativeGen)
	}

	// filling A
	a := make([]field.Element, n*r.Degree)
	ag := make([]field.Element, n*r.Degree)

	parallel.Execute(n, func(start, end int) {
		var buf bytes.Buffer
		for i := start; i < end; i++ {
			rstart, rend := i*r.Degree, (i+1)*r.Degree
			r.A[i] = a[rstart:rend:rend]
			r.Ag[i] = ag[rstart:rend:rend]
			for j := 0; j < r.Degree; j++ {
				r.A[i][j] = genRandom(seed, int64(i), int64(j), &buf)
			}

			// fill Ag the evaluation form of the polynomials in A on the coset √(g) * <g>
			copy(r.Ag[i], r.A[i])
			r.Domain.FFT(r.Ag[i], fft.DIF, fft.OnCoset())
		}
	})

	return r, nil
}

func (r *RSis) Write(p []byte) (n int, err error) {
	r.buffer.Write(p)
	return len(p), nil
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
// The instance buffer is interpreted as a sequence of coefficients of size r.Bound bits long.
// The function returns the hash of the polynomial as a a sequence []field.Elements, interpreted as []bytes,
// corresponding to sum_i A[i]*m Mod X^{d}+1
func (r *RSis) Sum(b []byte) []byte {
	buf := r.buffer.Bytes()
	if len(buf) > r.capacity {
		panic("buffer too large")
	}

	fastPath := r.LogTwoBound == 8 && r.Degree == 64

	// clear the buffers of the instance.
	defer r.cleanupBuffers()

	m := r.bufM
	mValues := r.bufMValues

	if fastPath {
		// fast path.
		limbDecomposeBytes8_64(buf, m, mValues)
	} else {
		limbDecomposeBytes(buf, m, r.LogTwoBound, r.Degree, mValues)
	}

	// we can hash now.
	res := r.bufRes

	// method 1: fft
	for i := 0; i < len(r.Ag); i++ {
		if !mValues.Test(uint(i)) {
			// means m[i*r.Degree : (i+1)*r.Degree] == [0...0]
			// we can skip this, FFT(0) = 0
			continue
		}
		k := m[i*r.Degree : (i+1)*r.Degree]
		if fastPath {
			// fast path.
			FFT64(k, r.twiddleCosets)
		} else {
			r.Domain.FFT(k, fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))
		}
		mulModAcc(res, r.Ag[i], k)
	}
	r.Domain.FFTInverse(res, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1)) // -> reduces mod Xᵈ+1

	resBytes, err := field.MarshalFieldVecToBinary(res)
	if err != nil {
		panic(err)
	}

	return append(b, resBytes[4:]...) // first 4 bytes are uint32(len(res))
}

// Reset resets the Hash to its initial state.
func (r *RSis) Reset() {
	r.buffer.Reset()
}

// Size returns the number of bytes Sum will return.
func (r *RSis) Size() int {

	// The size in bits is the size in bits of a polynomial in A.
	degree := len(r.A[0])
	totalSize := degree * field.Modulus().BitLen() / 8

	return totalSize
}

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
func (r *RSis) BlockSize() int {
	return 0
}

// Construct a hasher generator. It takes as input the same parameters
// as `NewRingSIS` and outputs a function which returns fresh hasher
// everytime it is called
func NewRingSISMaker(seed int64, logTwoDegree, logTwoBound, maxNbElementsToHash int) (func() hash.Hash, error) {
	return func() hash.Hash {
		h, err := NewRSis(seed, logTwoDegree, logTwoBound, maxNbElementsToHash)
		if err != nil {
			panic(err)
		}
		return h
	}, nil

}

func genRandom(seed, i, j int64, buf *bytes.Buffer) field.Element {

	buf.Reset()
	buf.WriteString("SIS")
	binary.Write(buf, binary.BigEndian, seed)
	binary.Write(buf, binary.BigEndian, i)
	binary.Write(buf, binary.BigEndian, j)

	digest := blake2b.Sum256(buf.Bytes())

	var res field.Element
	res.SetBytes(digest[:])

	return res
}

// mulMod computes p * q in ℤ_{p}[X]/Xᵈ+1.
// Is assumed that pLagrangeShifted and qLagrangeShifted are of the correct sizes
// and that they are in evaluation form on √(g) * <g>
// The result is not FFTinversed. The fft inverse is done once every
// multiplications are done.
func mulMod(pLagrangeCosetBitReversed, qLagrangeCosetBitReversed []field.Element) []field.Element {

	res := make([]field.Element, len(pLagrangeCosetBitReversed))
	for i := 0; i < len(pLagrangeCosetBitReversed); i++ {
		res[i].Mul(&pLagrangeCosetBitReversed[i], &qLagrangeCosetBitReversed[i])
	}

	// NOT fft inv for now, wait until every part of the keys have been multiplied
	// r.Domain.FFTInverse(res, fft.DIT, true)

	return res

}

// mulMod + accumulate in res.
func mulModAcc(res []field.Element, pLagrangeCosetBitReversed, qLagrangeCosetBitReversed []field.Element) {
	var t field.Element
	for i := 0; i < len(pLagrangeCosetBitReversed); i++ {
		t.Mul(&pLagrangeCosetBitReversed[i], &qLagrangeCosetBitReversed[i])
		res[i].Add(&res[i], &t)
	}
}

// Returns a clone of the RSis parameters with a fresh and empty buffer. Does not
// mutate the current instance. The keys and the public parameters of the SIS
// instance are not deep-copied. It is useful when we want to hash in parallel.
// Otherwise, we would have to generate an entire RSis for each thread.
func (r *RSis) CopyWithFreshBuffer() RSis {
	res := *r
	res.buffer = bytes.Buffer{}
	res.bufM = make([]field.Element, len(r.bufM))
	res.bufMValues = bitset.New(r.bufMValues.Len())
	res.bufRes = make([]field.Element, len(r.bufRes))
	return res
}

// Cleanup the buffers of the RSis instance
func (r *RSis) cleanupBuffers() {
	r.bufMValues.ClearAll()
	for i := 0; i < len(r.bufM); i++ {
		r.bufM[i].SetZero()
	}
	for i := 0; i < len(r.bufRes); i++ {
		r.bufRes[i].SetZero()
	}
}

// Split an slice of bytes representing an array of serialized field element in
// big-endian form into an array of limbs representing the same field elements
// in little-endian form. Namely, if our field is represented with 64 bits and we
// have the following field element 0x0123456789abcdef (0 being the most significant
// character and and f being the least significant one) and our log norm bound is
// 16 (so 1 hex character = 1 limb). The function assigns the values of m to [f, e,
// d, c, b, a, ..., 3, 2, 1, 0]. m should be preallocated and zeroized. Additionally,
// we have the guarantee that 2 bits contributing to different field elements cannot
// be part of the same limb.
func LimbDecomposeBytes(buf []byte, m []field.Element, logTwoBound int) {
	limbDecomposeBytes(buf, m, logTwoBound, 0, nil)
}

// Split an slice of bytes representing an array of serialized field element in
// big-endian form into an array of limbs representing the same field elements
// in little-endian form. Namely, if our field is represented with 64 bits and we
// have the following field element 0x0123456789abcdef (0 being the most significant
// character and and f being the least significant one) and our norm bound is
// 16 (so 1 hex character = 1 limb). The function assigns the values of m to [f, e,
// d, c, b, a, ..., 3, 2, 1, 0]. m should be preallocated and zeroized. mValues is
// an optional bitSet. If provided, it must be empty. The function will set bit "i"
// to indicate the that i-th SIS input polynomial should be non-zero. Recall, that a
// SIS polynomial corresponds to a chunk of limbs of size `degree`. Additionally,
// we have the guarantee that 2 bits contributing to different field elements cannot
// be part of the same limb.
func limbDecomposeBytes(buf []byte, m []field.Element, logTwoBound, degree int, mValues *bitset.BitSet) {

	// bitwise decomposition of the buffer, in order to build m (the vector to hash)
	// as a list of polynomials, whose coefficients are less than r.B bits long.
	// Say buf=[0xbe,0x0f]. As a stream of bits it is interpreted like this:
	// 10111110 00001111. BitAt(0)=1 (=leftmost bit), bitAt(1)=0 (=second leftmost bit), etc.
	nbBits := len(buf) * 8
	fieldBits := field.Bytes * 8

	// Decompose the buffer in blocks of logTwoBound bits, each forming a coefficient
	mPos := 0
	for fieldStart := 0; fieldStart < nbBits; fieldStart += fieldBits {

		// Use a single uint64 to hold the coefficient while collecting bits
		var coefficient uint64
		bitCount := 0
		for bitInField := 0; bitInField < fieldBits; {

			// Current bit position from the end of the field
			at := fieldStart + fieldBits - bitInField - 1

			// Check if we are within bounds of `buf`
			if k := at / 8; k < len(buf) {
				// Grab the bit from buf and shift it to the right place
				bit := (buf[k] >> (7 - (at % 8))) & 1
				coefficient |= uint64(bit) << (bitCount % logTwoBound)
				bitCount++
			}

			bitInField++

			// Check if we completed a coefficient (logTwoBound bits)
			if bitCount%logTwoBound == 0 || bitInField == fieldBits {
				m[mPos][0] = coefficient

				// Set bitset if coefficient is non-zero
				if coefficient != 0 && mValues != nil {
					mValues.Set(uint(mPos / degree))
				}

				// Reset coefficient and advance position
				coefficient = 0
				bitCount = 0
				mPos++
			}
		}
	}
}

// see limbDecomposeBytes; this function is optimized for the case where
// logTwoBound == 8 and degree == 64
func limbDecomposeBytes8_64(buf []byte, m []field.Element, mValues *bitset.BitSet) {
	// with logTwoBound == 8, we can actually advance byte per byte.
	const degree = 64
	j := 0

	for startPos := field.Bytes - 1; startPos < len(buf); startPos += field.Bytes {
		for i := startPos; i >= startPos-field.Bytes+1; i-- {
			m[j][0] = uint64(buf[i])
			if m[j][0] != 0 {
				mValues.Set(uint(j / degree))
			}
			j++
		}
	}
}
