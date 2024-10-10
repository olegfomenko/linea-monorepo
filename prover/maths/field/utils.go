package field

// This file is NOT autogenerated

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

func ToInt(e *Element) int {
	n := e.Uint64()
	if !e.IsUint64() || n > math.MaxInt {
		panic("out of range")
	}
	return int(n) // #nosec G115 -- Checked for overflow
}

// MarshalFieldVecToBinary marshals a slice of field elements to a binary representation.
func MarshalFieldVecToBinary(vec []Element) (data []byte, err error) {
	var w bytes.Buffer

	if err := binary.Write(&w, binary.BigEndian, uint32(len(vec))); err != nil {
		return nil, err
	}

	n := int64(4)

	var buf [Bytes]byte
	for i := 0; i < len(vec); i++ {
		fr.BigEndian.PutElement(&buf, vec[i])
		m, err := w.Write(buf[:])
		n += int64(m)
		if err != nil {
			return nil, err
		}
	}

	return w.Bytes(), nil
}
