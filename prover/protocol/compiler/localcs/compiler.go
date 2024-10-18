package localcs

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

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

/*
Compiles the local constraints
*/
func Compile(comp *wizard.CompiledIOP) {
	/*
		First compile all local constraints
	*/

	for _, qName := range comp.QueriesNoParams.AllUnignoredLocalConstraintKeys() {
		q := comp.QueriesNoParams.Data(qName).(query.LocalConstraint)
		ReduceLocalConstraint(comp, q, comp.QueriesNoParams.Round(qName))

	}
}
