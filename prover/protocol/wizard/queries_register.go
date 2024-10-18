package wizard

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
)

type QueriesRegistersProxy struct {
	LocalConstraint  ByRoundRegister[ifaces.QueryID, ifaces.Query]
	GlobalConstraint ByRoundRegister[ifaces.QueryID, ifaces.Query]
	LocalOpening     ByRoundRegister[ifaces.QueryID, ifaces.Query]
	InnerProduct     ByRoundRegister[ifaces.QueryID, ifaces.Query]
	Inclusion        ByRoundRegister[ifaces.QueryID, ifaces.Query]
	MiMC             ByRoundRegister[ifaces.QueryID, ifaces.Query]
	Permutation      ByRoundRegister[ifaces.QueryID, ifaces.Query]
	FixedPermutation ByRoundRegister[ifaces.QueryID, ifaces.Query]
	UnivariateEval   ByRoundRegister[ifaces.QueryID, ifaces.Query]
	Range            ByRoundRegister[ifaces.QueryID, ifaces.Query]
	All              ByRoundRegister[ifaces.QueryID, ifaces.Query]
}

func NewQueriesRegistersProxy() QueriesRegistersProxy {
	return QueriesRegistersProxy{
		LocalConstraint:  NewRegister[ifaces.QueryID, ifaces.Query](),
		GlobalConstraint: NewRegister[ifaces.QueryID, ifaces.Query](),
		LocalOpening:     NewRegister[ifaces.QueryID, ifaces.Query](),
		InnerProduct:     NewRegister[ifaces.QueryID, ifaces.Query](),
		Inclusion:        NewRegister[ifaces.QueryID, ifaces.Query](),
		MiMC:             NewRegister[ifaces.QueryID, ifaces.Query](),
		Permutation:      NewRegister[ifaces.QueryID, ifaces.Query](),
		FixedPermutation: NewRegister[ifaces.QueryID, ifaces.Query](),
		UnivariateEval:   NewRegister[ifaces.QueryID, ifaces.Query](),
		Range:            NewRegister[ifaces.QueryID, ifaces.Query](),
		All:              NewRegister[ifaces.QueryID, ifaces.Query](),
	}
}

func (r *QueriesRegistersProxy) AddToRound(round int, id ifaces.QueryID, data ifaces.Query) {
	r.All.AddToRound(round, id, data)
	switch data.(type) {
	case query.LocalConstraint:
		r.LocalConstraint.AddToRound(round, id, data)
	case query.GlobalConstraint:
		r.GlobalConstraint.AddToRound(round, id, data)
	case query.LocalOpening:
		r.LocalOpening.AddToRound(round, id, data)
	case query.InnerProduct:
		r.InnerProduct.AddToRound(round, id, data)
	case query.Inclusion:
		r.Inclusion.AddToRound(round, id, data)
	case query.MiMC:
		r.MiMC.AddToRound(round, id, data)
	case query.Permutation:
		r.Permutation.AddToRound(round, id, data)
	case query.FixedPermutation:
		r.FixedPermutation.AddToRound(round, id, data)
	case query.UnivariateEval:
		r.UnivariateEval.AddToRound(round, id, data)
	case query.Range:
		r.Range.AddToRound(round, id, data)
	}
}

func (r *QueriesRegistersProxy) MarkAsIgnored(id ifaces.QueryID) bool {
	data := r.All.Data(id)

	switch data.(type) {
	case query.LocalConstraint:
		r.LocalConstraint.MarkAsIgnored(id)
	case query.GlobalConstraint:
		r.GlobalConstraint.MarkAsIgnored(id)
	case query.LocalOpening:
		r.LocalOpening.MarkAsIgnored(id)
	case query.InnerProduct:
		r.InnerProduct.MarkAsIgnored(id)
	case query.Inclusion:
		r.Inclusion.MarkAsIgnored(id)
	case query.MiMC:
		r.MiMC.MarkAsIgnored(id)
	case query.Permutation:
		r.Permutation.MarkAsIgnored(id)
	case query.FixedPermutation:
		r.FixedPermutation.MarkAsIgnored(id)
	case query.UnivariateEval:
		r.UnivariateEval.MarkAsIgnored(id)
	case query.Range:
		r.Range.MarkAsIgnored(id)
	}

	return r.All.MarkAsIgnored(id)
}

func (r *QueriesRegistersProxy) IsIgnored(id ifaces.QueryID) bool {
	return r.All.IsIgnored(id)
}

func (r *QueriesRegistersProxy) ReserveFor(newLen int) {
	r.LocalConstraint.ReserveFor(newLen)
	r.GlobalConstraint.ReserveFor(newLen)
	r.LocalOpening.ReserveFor(newLen)
	r.InnerProduct.ReserveFor(newLen)
	r.Inclusion.ReserveFor(newLen)
	r.MiMC.ReserveFor(newLen)
	r.Permutation.ReserveFor(newLen)
	r.FixedPermutation.ReserveFor(newLen)
	r.UnivariateEval.ReserveFor(newLen)
	r.Range.ReserveFor(newLen)
	r.All.ReserveFor(newLen)
}
