package wizard

import (
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

/*
In a nutshell, an item is an abstract type that
accounts for the fact that CompiledProtocol
registers various things for different rounds
*/
type ByRoundRegister[ID comparable, DATA any] struct {
	// All the IDs with corresponding DATA
	mapping collection.Mapping[ID, DATA]
	// All the IDs for a given round
	byRounds collection.VecVec[ID]
	// Gives the round ID of an entry
	byRoundsIndex collection.Mapping[ID, int]
	// All the active (unignored) IDs for a given round
	byRoundsActive []collection.DynVec[ID]
	// Gives the position in byRoundsActive of an entry by ID
	byRoundsActiveIndex collection.Mapping[ID, uint]
}

/*
Construct a new round register
*/
func NewRegister[ID comparable, DATA any]() ByRoundRegister[ID, DATA] {
	return ByRoundRegister[ID, DATA]{
		mapping:             collection.NewMapping[ID, DATA](),
		byRounds:            collection.NewVecVec[ID](),
		byRoundsIndex:       collection.NewMapping[ID, int](),
		byRoundsActive:      []collection.DynVec[ID]{},
		byRoundsActiveIndex: collection.NewMapping[ID, uint](),
	}
}

/*
Insert for a given round. Will panic if an item
with the same ID has been registered first
*/
func (r *ByRoundRegister[ID, DATA]) AddToRound(round int, id ID, data DATA) {
	r.ReserveFor(round + 1)
	r.mapping.InsertNew(id, data)
	r.byRounds.AppendToInner(round, id)
	r.byRoundsIndex.InsertNew(id, round)
	r.byRoundsActiveIndex.InsertNew(id, r.byRoundsActive[round].Append(id))
}

/*
Returns the list of all the keys ever. The result is returned in
Deterministic order.
*/
func (r *ByRoundRegister[ID, DATA]) AllKeys() []ID {
	res := make([]ID, 0, r.mapping.Size())
	for roundID := 0; roundID < r.byRounds.Len(); roundID++ {
		ids := r.AllKeysAt(roundID)
		res = append(res, ids...)
	}
	return res
}

/*
Returns the list of all keys for a given round. Result has deterministic
order (order of insertion)
*/
func (r *ByRoundRegister[ID, DATA]) AllKeysAt(round int) []ID {
	// Reserve up to the desired length just in case.
	// It is absolutely legitimate to query "too far"
	// this can happens for queries for instance.
	// However, it should not happen for coins.
	r.byRounds.Reserve(round + 1)
	return r.byRounds.MustGet(round)
}

/*
Returns all the keys that are not marked as ignored in the structure
*/
func (r *ByRoundRegister[ID, DATA]) AllUnignoredKeys() []ID {
	res := make([]ID, 0, r.mapping.Size())
	for i := 0; i < r.byRounds.Len(); i++ {
		res = append(res, r.byRoundsActive[i].ListAll()...)
	}
	return res
}

/*
Returns the list of all unignored keys for a given round. Result has deterministic
order (order of insertion)
*/
func (r *ByRoundRegister[ID, DATA]) AllUnignoredKeysAt(round int) []ID {
	// Reserve up to the desired length just in case.
	// It is absolutely legitimate to query "too far"
	// this can happens for queries for instance.
	// However, it should not happen for coins.
	r.ReserveFor(round + 1)
	return r.byRoundsActive[round].ListAll()
}

/*
Returns the data for associated to an ID. Panic if not found
*/
func (r *ByRoundRegister[ID, DATA]) Data(id ID) DATA {
	return r.mapping.MustGet(id)
}

/*
Find
*/
func (r *ByRoundRegister[ID, DATA]) Round(id ID) int {
	return r.byRoundsIndex.MustGet(id)
}

/*
Panic if the name is not found at the given round
*/
func (r *ByRoundRegister[ID, DATA]) MustBeInRound(round int, id ID) {
	round_, ok := r.byRoundsIndex.TryGet(id)
	if !ok {
		utils.Panic("entry `%v` is not found at all. Was expecting to find it at round %v", id, round)
	}
	if round_ != round {
		utils.Panic("Wrong round, the entry %v was expected to be in round %v but found it in round %v", id, round, round_)
	}
}

/*
Panic if the name is not found at all
*/
func (r *ByRoundRegister[ID, DATA]) MustExists(id ...ID) {
	r.mapping.MustExists(id...)
}

/*
Returns true if all the entry exist
*/
func (r *ByRoundRegister[ID, DATA]) Exists(id ...ID) bool {
	return r.mapping.Exists(id...)
}

/*
Returns the number of rounds
*/
func (r *ByRoundRegister[ID, DATA]) NumRounds() int {
	return r.byRounds.Len()
}

/*
Make sure enough rounds are allocated up to the given length
No-op if enough rounds have been allocated, otherwise, will
reserve as many as necessary.
*/
func (r *ByRoundRegister[ID, DATA]) ReserveFor(newLen int) {
	if r.byRounds.Len() < newLen {
		r.byRounds.Reserve(newLen)
	}

	for len(r.byRoundsActive) < newLen {
		r.byRoundsActive = append(r.byRoundsActive, collection.NewDynVec[ID]())
	}
}

/*
Marks an entry as compiled. Panic if the key is missing from the register.
Returns true if the item was already ignored.
*/
func (r *ByRoundRegister[ID, DATA]) MarkAsIgnored(id ID) bool {
	round := r.byRoundsIndex.MustGet(id)

	index, active := r.byRoundsActiveIndex.TryGet(id)
	if !active {
		return true
	}

	r.byRoundsActive[round].MustRemove(index)
	r.byRoundsActiveIndex.Del(id)
	return false
}

/*
Returns if the entry is ignored. Panics if the entry is missing from the
map.
*/
func (r *ByRoundRegister[ID, DATA]) IsIgnored(id ID) bool {
	return r.byRoundsActiveIndex.Exists(id)
}
