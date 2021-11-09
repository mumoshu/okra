package analysis

import (
	"context"
	"fmt"
	"strings"
	"time"

	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
	"github.com/mumoshu/okra/pkg/clclient"
	"github.com/mumoshu/okra/pkg/okraerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RunInput struct {
	AnalysisTemplateName    string
	NS                      string
	AnalysisArgs            map[string]string
	AnalysisArgsFromSecrets map[string]string
	DryRun                  bool
}

// Run instantiates a new AnalysisRun object to let Argo Rollouts run an analysis.
// This command requires both AnalysisTemplate and AnalysisRun CRDs to be installed onto the cluster.
func Run(input RunInput) (*rolloutsv1alpha1.AnalysisRun, error) {
	c, err := clclient.New()
	if err != nil {
		return nil, okraerror.New(err)
	}

	templateName := input.AnalysisTemplateName
	ns := input.NS
	dryRun := input.DryRun

	var dryRunValues []string
	if dryRun {
		dryRunValues = []string{"All"}
	}

	ctx := context.Background()

	var template rolloutsv1alpha1.AnalysisTemplate
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: templateName}, &template); err != nil {
		return nil, okraerror.New(err)
	}

	argsMap := map[string]rolloutsv1alpha1.Argument{}
	for _, a := range template.Spec.Args {
		// This is the default value
		argsMap[a.Name] = a
	}

	// The following two sets of for-range loops is basically
	// an alternative to Argo Rollouts' MergeArgs.
	// We needed our own implementation here to deal with the fact that we can set
	// both args from immediate values and secretrefs.
	// See below for MergeArgs
	// https://github.com/argoproj/argo-rollouts/blob/1ee46cff2a3203fd2da7d540c9fd25c8a61900c2/utils/analysis/helpers.go#L165-L167

	for k, v := range input.AnalysisArgs {
		if _, ok := argsMap[k]; !ok {
			return nil, okraerror.New(fmt.Errorf("argument %s does not exist in analysisrun template %s", k, templateName))
		}

		v := v

		argsMap[k] = rolloutsv1alpha1.Argument{
			Name:  k,
			Value: &v,
		}
	}

	for k, v := range input.AnalysisArgsFromSecrets {
		if _, ok := argsMap[k]; !ok {
			return nil, okraerror.New(fmt.Errorf("argument %s does not exist in analysisrun template %s", k, templateName))
		}

		vs := strings.SplitN(v, ".", 2)

		argsMap[k] = rolloutsv1alpha1.Argument{
			Name: k,
			ValueFrom: &rolloutsv1alpha1.ValueFrom{
				SecretKeyRef: &rolloutsv1alpha1.SecretKeyRef{
					Name: vs[0],
					Key:  vs[1],
				},
			},
		}
	}

	var args []rolloutsv1alpha1.Argument

	for _, v := range argsMap {
		args = append(args, v)
	}

	timestamp := time.Now().Format("20060102150405")

	const TimestampLabel = "okra.mumo.co/timestamp"

	runLabels := map[string]string{
		TimestampLabel: timestamp,
	}

	run := rolloutsv1alpha1.AnalysisRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: templateName + "-",
			Namespace:    ns,
			Labels:       runLabels,
		},
		Spec: rolloutsv1alpha1.AnalysisRunSpec{
			Args:    args,
			Metrics: template.Spec.Metrics,
		},
	}

	if err := c.Create(ctx, &run, &client.CreateOptions{DryRun: dryRunValues}); err != nil {
		return nil, okraerror.New(err)
	}

	var created rolloutsv1alpha1.AnalysisRunList

	var opts []client.ListOption

	if ns != "" {
		opts = append(opts, client.InNamespace(ns))
	}

	lbls, err := labels.ValidatedSelectorFromSet(runLabels)
	if err != nil {
		return nil, okraerror.New(err)
	}

	opts = append(opts, client.MatchingLabelsSelector{Selector: lbls})

	if err := c.List(ctx, &created, opts...); err != nil {
		return nil, okraerror.New(err)
	}

	if len(created.Items) != 1 {
		return nil, okraerror.New(fmt.Errorf("unexpected number of runs found: %d", len(created.Items)))
	}

	return &created.Items[0], nil
}
