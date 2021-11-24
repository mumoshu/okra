package cmd

import (
	"fmt"
	"os"

	"github.com/mumoshu/okra/pkg/analysis"
	"github.com/spf13/cobra"
)

func createAnalysisRunCommand() *cobra.Command {
	var c analysis.RunInput

	cmd := &cobra.Command{
		Use: "analysisrun",
		RunE: func(cmd *cobra.Command, args []string) error {
			run, err := analysis.Run(c)

			if run != nil {
				fmt.Fprintf(os.Stdout, "%+v\n", *run)
			}

			return err
		},
	}

	flag := cmd.Flags()

	flag.StringVar(&c.AnalysisTemplateName, "template-name", "", "")
	flag.StringVar(&c.NS, "namespace", "", "")
	flag.StringToStringVar(&c.AnalysisArgs, "args", map[string]string{}, "")
	flag.StringToStringVar(&c.AnalysisArgsFromSecrets, "args-from-secrets", map[string]string{}, "A list of secret refs like \"arg-name=secret-name.field-name\" concatenated by \",\"s")

	return cmd
}
