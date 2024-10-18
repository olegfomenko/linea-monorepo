package specialqueries

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

/*
Reduce the fixed permutations into
*/
func CompileFixedPermutations(comp *wizard.CompiledIOP) {
	for _, qName := range comp.QueriesNoParams.AllUnignoredFixedPermutationKeys() {
		q := comp.QueriesNoParams.Data(qName).(query.FixedPermutation)
		reduceFixedPermutation(comp, q, comp.QueriesNoParams.Round(qName))
	}
}

/*
Derive a name for a a coin created during the compilation process
*/
func deriveName[R ~string](context string, q ifaces.QueryID, name string) R {
	var res string
	if q == "" {
		res = fmt.Sprintf("%v_%v", context, name)
	} else {
		res = fmt.Sprintf("%v_%v_%v", q, context, name)
	}
	return R(res)
}
