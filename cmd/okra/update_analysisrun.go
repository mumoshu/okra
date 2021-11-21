package okra

import (
	"fmt"
	"strings"

	"github.com/mumoshu/okra/pkg/analysis"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
)

func updateAnalysisRunCommand() *cobra.Command {
	var input func() *analysis.UpdateInput
	cmd := &cobra.Command{
		Use: "update-analysisrun",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := analysis.Update(*input())
			return err
		},
	}
	input = updateAnalysisRunFlags(cmd.Flags(), &analysis.UpdateInput{})
	return cmd
}

func updateAnalysisRunFlags(flag *pflag.FlagSet, c *analysis.UpdateInput) func() *analysis.UpdateInput {
	var phase string

	flag.StringVar(&c.NS, "namespace", "", "Namespace of the analysis run")
	flag.StringVar(&c.Name, "name", "", "Name of the analysis run")
	flag.StringVar(&phase, "phase", string(rolloutsv1alpha1.AnalysisPhaseSuccessful), "Phase of the analysis run")

	return func() *analysis.UpdateInput {
		var p rolloutsv1alpha1.AnalysisPhase

		lowerCased := strings.ToLower(phase)

		switch lowerCased {
		case "successful":
			p = rolloutsv1alpha1.AnalysisPhaseSuccessful
		case "error":
			p = rolloutsv1alpha1.AnalysisPhaseError
		case "failed":
			p = rolloutsv1alpha1.AnalysisPhaseFailed
		case "inconclusive":
			p = rolloutsv1alpha1.AnalysisPhaseInconclusive
		default:
			panic(fmt.Errorf("unsupported phase: %s", phase))
		}

		input := c
		input.Phase = p

		return input
	}
}
