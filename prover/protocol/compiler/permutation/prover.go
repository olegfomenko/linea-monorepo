package permutation

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// proverTaskAtRound implements the [wizard.ProverAction] interface and is
// responsible for assigning the Z polynomials of all the queries for which the
// Z polynomial needs to be assigned in the current round
type proverTaskAtRound []*ZCtx

func (p proverTaskAtRound) Run(run *wizard.ProverRuntime) {
	parallel.Execute(len(p), func(start, end int) {
		for i := start; i < end; i++ {
			p[i].run(run)
		}
	})
}

// run assigns all the Zs in parallel and set the parameters for their
// corresponding last values openings.
func (z *ZCtx) run(run *wizard.ProverRuntime) {

	for i := range z.Zs {
		var (
			numerator   []field.Element
			denominator []field.Element
		)

		if packingArity*i < len(z.NumeratorFactors) {
			numerator = wizardutils.EvalExprColumn(run, z.NumeratorFactorsBoarded[i]).IntoRegVecSaveAlloc()
		} else {
			numerator = vector.Repeat(field.One(), z.Size)
		}

		if packingArity*i < len(z.DenominatorFactors) {
			denominator = wizardutils.EvalExprColumn(run, z.DenominatorFactorsBoarded[i]).IntoRegVecSaveAlloc()
		} else {
			denominator = vector.Repeat(field.One(), z.Size)
		}

		denominator = field.BatchInvert(denominator)

		for i := range denominator {
			numerator[i].Mul(&numerator[i], &denominator[i])
			if i > 0 {
				numerator[i].Mul(&numerator[i], &numerator[i-1])
			}
		}

		run.AssignColumn(z.Zs[i].GetColID(), smartvectors.NewRegular(numerator))
		run.AssignLocalPoint(z.ZOpenings[i].Name(), numerator[len(numerator)-1])
	}

}
