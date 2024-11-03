package innerproduct

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// proverTask implements the [wizard.ProverAction] interface and as such
// implements the prover work of the compilation step. It works by calling
// in parallel the prover tasks of the sub-compilation steps.
type proverTask []*contextForSize

// Run implements the [wizard.ProverAction] interface.
func (p proverTask) Run(run *wizard.ProverRuntime) {
	parallel.Execute(len(p), func(start, end int) {
		for i := start; i < end; i++ {
			p[i].run(run)
		}
	})
}

// run partially implements the prover runtime associated with the current
// partial compilation context. Its role is to assign Summation and its opening.
func (ctx *contextForSize) run(run *wizard.ProverRuntime) {

	var (
		size      = ctx.Summation.Size()
		collapsed = wizardutils.EvalExprColumn(run, ctx.CollapsedBoard).IntoRegVecSaveAlloc()
		summation = make([]field.Element, size)
	)

	summation[0] = collapsed[0]

	for i := 0; i+1 < size; i++ {
		summation[i+1].Add(&summation[i], &collapsed[i+1])
	}

	run.AssignColumn(ctx.Summation.GetColID(), smartvectors.NewRegular(summation))
	run.AssignLocalPoint(ctx.SummationOpening.ID, summation[len(summation)-1])
}
