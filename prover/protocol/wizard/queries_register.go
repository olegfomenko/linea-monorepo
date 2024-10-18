package wizard

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

type QueryByRoundRegister struct {
	// All the data for each key
	mapping collection.Mapping[ifaces.QueryID, ifaces.Query]
	// Gives the round ID of an entry
	byRoundsIndex collection.Mapping[ifaces.QueryID, int]
	// Marks an entry as ignorable (but does not delete it)
	ignored collection.Set[ifaces.QueryID]

	// All the IDs for a given round
	all              collection.VecVec[ifaces.QueryID]
	localConstraint  collection.VecVec[ifaces.QueryID]
	globalConstraint collection.VecVec[ifaces.QueryID]
	localOpening     collection.VecVec[ifaces.QueryID]
	innerProduct     collection.VecVec[ifaces.QueryID]
	inclusion        collection.VecVec[ifaces.QueryID]
	miMC             collection.VecVec[ifaces.QueryID]
	permutation      collection.VecVec[ifaces.QueryID]
	fixedPermutation collection.VecVec[ifaces.QueryID]
	univariateEval   collection.VecVec[ifaces.QueryID]
	rangee           collection.VecVec[ifaces.QueryID]
}

func NewQueriesRegistersProxy() QueryByRoundRegister {
	return QueryByRoundRegister{
		mapping:          collection.NewMapping[ifaces.QueryID, ifaces.Query](),
		byRoundsIndex:    collection.NewMapping[ifaces.QueryID, int](),
		ignored:          collection.NewSet[ifaces.QueryID](),
		all:              collection.NewVecVec[ifaces.QueryID](),
		localConstraint:  collection.NewVecVec[ifaces.QueryID](),
		globalConstraint: collection.NewVecVec[ifaces.QueryID](),
		localOpening:     collection.NewVecVec[ifaces.QueryID](),
		innerProduct:     collection.NewVecVec[ifaces.QueryID](),
		inclusion:        collection.NewVecVec[ifaces.QueryID](),
		miMC:             collection.NewVecVec[ifaces.QueryID](),
		permutation:      collection.NewVecVec[ifaces.QueryID](),
		fixedPermutation: collection.NewVecVec[ifaces.QueryID](),
		univariateEval:   collection.NewVecVec[ifaces.QueryID](),
		rangee:           collection.NewVecVec[ifaces.QueryID](),
	}
}

func (r *QueryByRoundRegister) AddToRound(round int, id ifaces.QueryID, data ifaces.Query) {
	r.mapping.InsertNew(id, data)
	r.byRoundsIndex.InsertNew(id, round)

	r.all.AppendToInner(round, id)
	switch data.(type) {
	case query.LocalConstraint:
		r.localConstraint.AppendToInner(round, id)
	case query.GlobalConstraint:
		r.globalConstraint.AppendToInner(round, id)
	case query.LocalOpening:
		r.localOpening.AppendToInner(round, id)
	case query.InnerProduct:
		r.innerProduct.AppendToInner(round, id)
	case query.Inclusion:
		r.inclusion.AppendToInner(round, id)
	case query.MiMC:
		r.miMC.AppendToInner(round, id)
	case query.Permutation:
		r.permutation.AppendToInner(round, id)
	case query.FixedPermutation:
		r.fixedPermutation.AppendToInner(round, id)
	case query.UnivariateEval:
		r.univariateEval.AppendToInner(round, id)
	case query.Range:
		r.rangee.AppendToInner(round, id)
	}
}

func (r *QueryByRoundRegister) AllKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, r.all.Count())
	for roundID := 0; roundID < r.NumRounds(); roundID++ {
		ids := r.AllKeysAt(roundID)
		res = append(res, ids...)
	}
	return res
}

func (r *QueryByRoundRegister) AllLocalConstraintKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, r.localConstraint.Count())
	for roundID := 0; roundID < r.NumRounds(); roundID++ {
		ids := r.AllLocalConstraintKeysAt(roundID)
		res = append(res, ids...)
	}
	return res
}

func (r *QueryByRoundRegister) AllGlobalConstraintKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, r.globalConstraint.Count())
	for roundID := 0; roundID < r.NumRounds(); roundID++ {
		ids := r.AllGlobalConstraintKeysAt(roundID)
		res = append(res, ids...)
	}
	return res
}

func (r *QueryByRoundRegister) AllLocalOpeningKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, r.localOpening.Count())
	for roundID := 0; roundID < r.NumRounds(); roundID++ {
		ids := r.AllLocalOpeningKeysAt(roundID)
		res = append(res, ids...)
	}
	return res
}

func (r *QueryByRoundRegister) AllInnerProductKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, r.innerProduct.Count())
	for roundID := 0; roundID < r.NumRounds(); roundID++ {
		ids := r.AllInnerProductKeysAt(roundID)
		res = append(res, ids...)
	}
	return res
}

func (r *QueryByRoundRegister) AllInclusionKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, r.inclusion.Count())
	for roundID := 0; roundID < r.NumRounds(); roundID++ {
		ids := r.AllInclusionKeysAt(roundID)
		res = append(res, ids...)
	}
	return res
}

func (r *QueryByRoundRegister) AllMiMCKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, r.miMC.Count())
	for roundID := 0; roundID < r.NumRounds(); roundID++ {
		ids := r.AllMiMCKeysAt(roundID)
		res = append(res, ids...)
	}
	return res
}

func (r *QueryByRoundRegister) AllPermutationKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, r.permutation.Count())
	for roundID := 0; roundID < r.NumRounds(); roundID++ {
		ids := r.AllPermutationKeysAt(roundID)
		res = append(res, ids...)
	}
	return res
}

func (r *QueryByRoundRegister) AllFixedPermutationKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, r.fixedPermutation.Count())
	for roundID := 0; roundID < r.NumRounds(); roundID++ {
		ids := r.AllFixedPermutationKeysAt(roundID)
		res = append(res, ids...)
	}
	return res
}

func (r *QueryByRoundRegister) AllUnivariateEvalKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, r.univariateEval.Count())
	for roundID := 0; roundID < r.NumRounds(); roundID++ {
		ids := r.AllUnivariateEvalKeysAt(roundID)
		res = append(res, ids...)
	}
	return res
}

func (r *QueryByRoundRegister) AllRangeKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, r.rangee.Count())
	for roundID := 0; roundID < r.NumRounds(); roundID++ {
		ids := r.AllRangeKeysAt(roundID)
		res = append(res, ids...)
	}
	return res
}

func (r *QueryByRoundRegister) AllKeysAt(round int) []ifaces.QueryID {
	r.ReserveFor(round + 1)
	return r.all.MustGet(round)
}

func (r *QueryByRoundRegister) AllLocalConstraintKeysAt(round int) []ifaces.QueryID {
	r.ReserveFor(round + 1)
	return r.localConstraint.MustGet(round)
}

func (r *QueryByRoundRegister) AllGlobalConstraintKeysAt(round int) []ifaces.QueryID {
	r.ReserveFor(round + 1)
	return r.globalConstraint.MustGet(round)
}

func (r *QueryByRoundRegister) AllLocalOpeningKeysAt(round int) []ifaces.QueryID {
	r.ReserveFor(round + 1)
	return r.localOpening.MustGet(round)
}

func (r *QueryByRoundRegister) AllInnerProductKeysAt(round int) []ifaces.QueryID {
	r.ReserveFor(round + 1)
	return r.innerProduct.MustGet(round)
}

func (r *QueryByRoundRegister) AllInclusionKeysAt(round int) []ifaces.QueryID {
	r.ReserveFor(round + 1)
	return r.inclusion.MustGet(round)
}

func (r *QueryByRoundRegister) AllMiMCKeysAt(round int) []ifaces.QueryID {
	r.ReserveFor(round + 1)
	return r.miMC.MustGet(round)
}

func (r *QueryByRoundRegister) AllPermutationKeysAt(round int) []ifaces.QueryID {
	r.ReserveFor(round + 1)
	return r.permutation.MustGet(round)
}

func (r *QueryByRoundRegister) AllFixedPermutationKeysAt(round int) []ifaces.QueryID {
	r.ReserveFor(round + 1)
	return r.fixedPermutation.MustGet(round)
}

func (r *QueryByRoundRegister) AllUnivariateEvalKeysAt(round int) []ifaces.QueryID {
	r.ReserveFor(round + 1)
	return r.univariateEval.MustGet(round)
}

func (r *QueryByRoundRegister) AllRangeKeysAt(round int) []ifaces.QueryID {
	r.ReserveFor(round + 1)
	return r.rangee.MustGet(round)
}

func (r *QueryByRoundRegister) Data(id ifaces.QueryID) ifaces.Query {
	return r.mapping.MustGet(id)
}

func (r *QueryByRoundRegister) Round(id ifaces.QueryID) int {
	round, ok := r.byRoundsIndex.TryGet(id)
	if !ok {
		utils.Panic("Could not find entry %v", id)
	}
	return round
}

func (r *QueryByRoundRegister) MustBeInRound(round int, id ifaces.QueryID) {
	round_, ok := r.byRoundsIndex.TryGet(id)
	if !ok {
		utils.Panic("entry `%v` is not found at all. Was expecting to find it at round %v", id, round)
	}
	if round_ != round {
		utils.Panic("Wrong round, the entry %v was expected to be in round %v but found it in round %v", id, round, round_)
	}
}

func (r *QueryByRoundRegister) MustExists(id ...ifaces.QueryID) {
	r.mapping.MustExists(id...)
}

func (r *QueryByRoundRegister) Exists(id ...ifaces.QueryID) bool {
	return r.mapping.Exists(id...)
}

func (r *QueryByRoundRegister) NumRounds() int {
	return r.all.Len()
}

func (r *QueryByRoundRegister) ReserveFor(newLen int) {
	r.all.Reserve(newLen)
	r.localConstraint.Reserve(newLen)
	r.globalConstraint.Reserve(newLen)
	r.localOpening.Reserve(newLen)
	r.innerProduct.Reserve(newLen)
	r.inclusion.Reserve(newLen)
	r.miMC.Reserve(newLen)
	r.permutation.Reserve(newLen)
	r.fixedPermutation.Reserve(newLen)
	r.univariateEval.Reserve(newLen)
	r.rangee.Reserve(newLen)
}

func (r *QueryByRoundRegister) AllUnignoredKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, r.ignored.Size())

	for i := 0; i < r.NumRounds(); i++ {
		allKeys := r.AllKeysAt(i)
		for _, k := range allKeys {
			if r.IsIgnored(k) {
				continue
			}
			res = append(res, k)
		}
	}
	return res
}

func (r *QueryByRoundRegister) AllUnignoredLocalConstraintKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, min(r.ignored.Size(), r.localConstraint.Count()))

	for i := 0; i < r.NumRounds(); i++ {
		allKeys := r.AllLocalConstraintKeysAt(i)
		for _, k := range allKeys {
			if r.IsIgnored(k) {
				continue
			}
			res = append(res, k)
		}
	}
	return res
}

func (r *QueryByRoundRegister) AllUnignoredGlobalConstraintKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, min(r.ignored.Size(), r.globalConstraint.Count()))

	for i := 0; i < r.NumRounds(); i++ {
		allKeys := r.AllGlobalConstraintKeysAt(i)
		for _, k := range allKeys {
			if r.IsIgnored(k) {
				continue
			}
			res = append(res, k)
		}
	}
	return res
}

func (r *QueryByRoundRegister) AllUnignoredLocalOpeningKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, min(r.ignored.Size(), r.localOpening.Count()))

	for i := 0; i < r.NumRounds(); i++ {
		allKeys := r.AllLocalOpeningKeysAt(i)
		for _, k := range allKeys {
			if r.IsIgnored(k) {
				continue
			}
			res = append(res, k)
		}
	}
	return res
}

func (r *QueryByRoundRegister) AllUnignoredInnerProductKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, min(r.ignored.Size(), r.innerProduct.Count()))

	for i := 0; i < r.NumRounds(); i++ {
		allKeys := r.AllInnerProductKeysAt(i)
		for _, k := range allKeys {
			if r.IsIgnored(k) {
				continue
			}
			res = append(res, k)
		}
	}
	return res
}

func (r *QueryByRoundRegister) AllUnignoredInclusionKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, min(r.ignored.Size(), r.inclusion.Count()))

	for i := 0; i < r.NumRounds(); i++ {
		allKeys := r.AllInclusionKeysAt(i)
		for _, k := range allKeys {
			if r.IsIgnored(k) {
				continue
			}
			res = append(res, k)
		}
	}
	return res
}

func (r *QueryByRoundRegister) AllUnignoredMiMCKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, min(r.ignored.Size(), r.miMC.Count()))

	for i := 0; i < r.NumRounds(); i++ {
		allKeys := r.AllMiMCKeysAt(i)
		for _, k := range allKeys {
			if r.IsIgnored(k) {
				continue
			}
			res = append(res, k)
		}
	}
	return res
}

func (r *QueryByRoundRegister) AllUnignoredPermutationKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, min(r.ignored.Size(), r.permutation.Count()))

	for i := 0; i < r.NumRounds(); i++ {
		allKeys := r.AllPermutationKeysAt(i)
		for _, k := range allKeys {
			if r.IsIgnored(k) {
				continue
			}
			res = append(res, k)
		}
	}
	return res
}

func (r *QueryByRoundRegister) AllUnignoredFixedPermutationKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, min(r.ignored.Size(), r.fixedPermutation.Count()))

	for i := 0; i < r.NumRounds(); i++ {
		allKeys := r.AllFixedPermutationKeysAt(i)
		for _, k := range allKeys {
			if r.IsIgnored(k) {
				continue
			}
			res = append(res, k)
		}
	}
	return res
}

func (r *QueryByRoundRegister) AllUnignoredUnivariateEvalKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, min(r.ignored.Size(), r.univariateEval.Count()))

	for i := 0; i < r.NumRounds(); i++ {
		allKeys := r.AllUnivariateEvalKeysAt(i)
		for _, k := range allKeys {
			if r.IsIgnored(k) {
				continue
			}
			res = append(res, k)
		}
	}
	return res
}

func (r *QueryByRoundRegister) AllUnignoredRangeKeys() []ifaces.QueryID {
	res := make([]ifaces.QueryID, 0, min(r.ignored.Size(), r.rangee.Count()))

	for i := 0; i < r.NumRounds(); i++ {
		allKeys := r.AllRangeKeysAt(i)
		for _, k := range allKeys {
			if r.IsIgnored(k) {
				continue
			}
			res = append(res, k)
		}
	}
	return res
}

func (r *QueryByRoundRegister) MarkAsIgnored(id ifaces.QueryID) bool {
	r.mapping.MustExists(id)
	return r.ignored.Insert(id)
}

func (r *QueryByRoundRegister) IsIgnored(id ifaces.QueryID) bool {
	r.mapping.MustExists(id)
	return r.ignored.Exists(id)
}
