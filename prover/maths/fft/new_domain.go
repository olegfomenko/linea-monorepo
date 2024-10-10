package fft

import (
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

/*
Creates a domain without a coset
*/
func NewDomain(m int) *Domain {

	// Sanity-check
	if !utils.IsPowerOfTwo(m) {
		utils.Panic("`m` is not a power of two %v", m)
	}

	// Sanity-check
	if m > 1<<maxOrderInt {
		utils.Panic("The current field does not have a `m`-roots of unity group (m = %v)", m)
	}

	domain := &Domain{}
	order := utils.Log2Ceil(m)
	domain.Cardinality = uint64(m)

	// Multiplicative generator of FF* (not a 2-adic root of unity)
	domain.FrMultiplicativeGen.SetUint64(field.MultiplicativeGen)
	domain.FrMultiplicativeGenInv.Inverse(&domain.FrMultiplicativeGen)

	// Generator = FinerGenerator^2 has order x
	expo := uint64(1 << (maxOrderInt - order))
	var expoBig big.Int
	expoBig.SetUint64(expo)
	// order x
	domain.Generator.Exp(field.RootOfUnity, &expoBig)
	domain.GeneratorInv.Inverse(&domain.Generator)
	domain.CardinalityInv.SetUint64(uint64(m)).Inverse(&domain.CardinalityInv)

	// Either get the twiddles or recompute them
	domain.Twiddles, domain.TwiddlesInv = GetTwiddleForDomainOfSize(m)
	return domain
}

/*
Equip the current domain with a coset shifted by the multiplicative generator
*/
func (dom *Domain) WithCoset() *Domain {
	return dom.WithCustomCoset(1, 0)
}

/*
Equipe the current domain with a custom coset obtained as explained in
the doc of `GetCoset`
*/
func (dom *Domain) WithCustomCoset(r, numcoset int) *Domain {
	n := utils.ToInt(dom.Cardinality)
	dom.CosetTable,
		dom.CosetTableInv,
		dom.CosetTableReversed,
		dom.CosetTableInvReversed = GetCoset(n, r, numcoset)

	return dom
}

/*
WithShift sets the FrMultiplicativeGen of the domain.
Default is generator of the largest 2-adic subgroup.

And equipes the current domain with a custom coset generated
with shifted FrMultiplicativeGen.
*/
func (dom *Domain) WithShift(shift field.Element) *Domain {
	dom.FrMultiplicativeGen.Set(&shift)
	dom.FrMultiplicativeGenInv.Inverse(&dom.FrMultiplicativeGen)

	dom.CosetTable = make([]field.Element, dom.Cardinality)
	dom.CosetTableInv = make([]field.Element, dom.Cardinality)

	var wg sync.WaitGroup

	expTable := func(sqrt field.Element, t []field.Element) {
		fft.BuildExpTable(sqrt, t)
		wg.Done()
	}

	wg.Add(2)
	go expTable(dom.FrMultiplicativeGen, dom.CosetTable)
	go expTable(dom.FrMultiplicativeGenInv, dom.CosetTableInv)

	wg.Wait()

	return dom
}
