package wizard

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

/*
In a nutshell, an item is an abstract type that
accounts for the fact that CompiledProtocol
registers various things for different rounds
*/
type CoinByRoundRegister struct {
	// All the data for each key
	mapping collection.Mapping[coin.Name, coin.Info]
	// All the IDs for a given round
	byRounds collection.VecVec[coin.Name]
}

/*
Construct a new round register
*/
func NewCoinByRoundRegister() CoinByRoundRegister {
	return CoinByRoundRegister{
		mapping:  collection.NewMapping[coin.Name, coin.Info](),
		byRounds: collection.NewVecVec[coin.Name](),
	}
}

/*
Insert for a given round. Will panic if an item
with the same ID has been registered first
*/
func (r *CoinByRoundRegister) AddToRound(round int, id coin.Name, data coin.Info) {
	r.mapping.InsertNew(id, data)
	r.byRounds.AppendToInner(round, id)
}

/*
Returns the list of all keys for a given round. Result has deterministic
order (order of insertion)
*/
func (r *CoinByRoundRegister) AllKeysAt(round int) []coin.Name {
	// Reserve up to the desired length just in case.
	// It is absolutely legitimate to query "too far"
	// this can happens for queries for instance.
	// However, it should not happen for coins.
	r.byRounds.Reserve(round + 1)
	return r.byRounds.MustGet(round)
}

/*
Returns the data for associated to an ID. Panic if not found
*/
func (r *CoinByRoundRegister) Data(id coin.Name) coin.Info {
	return r.mapping.MustGet(id)
}

/*
Find
*/
func (r *CoinByRoundRegister) Round(id coin.Name) int {
	coin := r.mapping.MustGet(id)
	return coin.Round
}

/*
Panic if the name is not found at the given round
*/
func (r *CoinByRoundRegister) MustBeInRound(round int, id coin.Name) {
	if round_ := r.Round(id); round_ != round {
		utils.Panic("Wrong round, the entry %v was expected to be in round %v but found it in round %v", id, round, round_)
	}
}

/*
Panic if the name is not found at all
*/
func (r *CoinByRoundRegister) MustExists(id ...coin.Name) {
	r.mapping.MustExists(id...)
}

/*
Returns the number of rounds
*/
func (r *CoinByRoundRegister) NumRounds() int {
	return r.byRounds.Len()
}

/*
Make sure enough rounds are allocated up to the given length
No-op if enough rounds have been allocated, otherwise, will
reserve as many as necessary.
*/
func (r *CoinByRoundRegister) ReserveFor(newLen int) {
	if r.byRounds.Len() < newLen {
		r.byRounds.Reserve(newLen)
	}
}
