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
	// All the IDs for a given round
	byRounds []collection.Mapping[ID, DATA]
	// All the unignored IDs for a given round
	byRoundsUnignored []collection.Set[ID]
	// Gives the round ID of an entry
	byRoundsIndex collection.Mapping[ID, int]
}

/*
Construct a new round register
*/
func NewRegister[ID comparable, DATA any]() ByRoundRegister[ID, DATA] {
	return ByRoundRegister[ID, DATA]{
		byRounds:          []collection.Mapping[ID, DATA]{},
		byRoundsIndex:     collection.NewMapping[ID, int](),
		byRoundsUnignored: []collection.Set[ID]{},
	}
}

/*
Insert for a given round. Will panic if an item
with the same ID has been registered first
*/
func (r *ByRoundRegister[ID, DATA]) AddToRound(round int, id ID, data DATA) {
	r.ReserveFor(round + 1)
	r.byRounds[round].InsertNew(id, data)
	r.byRoundsUnignored[round].InsertNew(id)
	r.byRoundsIndex.InsertNew(id, round)
}

/*
Returns the list of all the keys ever. The result is returned in
Deterministic order.
*/
func (r *ByRoundRegister[ID, DATA]) AllKeys() []ID {
	res := []ID{}
	for i := 0; i < len(r.byRounds); i++ {
		res = append(res, r.byRounds[i].ListAllKeys()...)
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
	r.ReserveFor(round + 1)
	return r.byRounds[round].ListAllKeys()
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
	return r.byRoundsUnignored[round].ListAll()
}

/*
Returns the data for associated to an ID. Panic if not found
*/
func (r *ByRoundRegister[ID, DATA]) Data(id ID) DATA {
	round := r.byRoundsIndex.MustGet(id)
	return r.byRounds[round].MustGet(id)
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
	r.byRoundsIndex.MustExists(id...)
}

/*
Returns true if all the entry exist
*/
func (r *ByRoundRegister[ID, DATA]) Exists(id ...ID) bool {
	return r.byRoundsIndex.Exists(id...)
}

/*
Returns the number of rounds
*/
func (r *ByRoundRegister[ID, DATA]) NumRounds() int {
	return len(r.byRounds)
}

/*
Make sure enough rounds are allocated up to the given length
No-op if enough rounds have been allocated, otherwise, will
reserve as many as necessary.
*/
func (r *ByRoundRegister[ID, DATA]) ReserveFor(newLen int) {
	for len(r.byRounds) < newLen {
		r.byRounds = append(r.byRounds, collection.NewMapping[ID, DATA]())
		r.byRoundsUnignored = append(r.byRoundsUnignored, collection.NewSet[ID]())
	}
}

/*
Returns all the keys that are not marked as ignored in the structure
*/
func (s *ByRoundRegister[ID, DATA]) AllUnignoredKeys() []ID {
	res := []ID{}
	for r := 0; r < len(s.byRounds); r++ {
		res = append(res, s.byRoundsUnignored[r].ListAll()...)
	}
	return res
}

/*
Marks an entry as compiled. Panic if the key is missing from the register.
Returns true if the item was already ignored.
*/
func (r *ByRoundRegister[ID, DATA]) MarkAsIgnored(id ID) bool {
	round := r.byRoundsIndex.MustGet(id)

	ignored := !r.byRoundsUnignored[round].Exists(id)
	if !ignored {
		r.byRoundsUnignored[round].MustDel(id)
	}

	return ignored
}

/*
Returns if the entry is ignored. Panics if the entry is missing from the
map.
*/
func (r *ByRoundRegister[ID, DATA]) IsIgnored(id ID) bool {
	round := r.byRoundsIndex.MustGet(id)
	return r.byRoundsUnignored[round].Exists(id)
}
