package okra

import (
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/aws/aws-sdk-go/service/eks"
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/analysis"
	"github.com/mumoshu/okra/pkg/awstargetgroupset"
	"github.com/mumoshu/okra/pkg/clusterset"
	"github.com/mumoshu/okra/pkg/manager"
	"github.com/mumoshu/okra/pkg/okraerror"
	"github.com/mumoshu/okra/pkg/targetgroupbinding"
	_ "k8s.io/client-go/plugin/pkg/client/auth/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	ApplicationName = "okra"
)

type Config struct {
	dryRun   bool
	ns       string
	name     string
	endpoint string
	caData   string
	eksTags  []string
	labelKVs []string
}

func InitFlags(flag *pflag.FlagSet) *Config {
	var c Config

	flag.BoolVar(&c.dryRun, "dry-run", false, "")
	flag.StringVar(&c.ns, "namespace", "", "")
	flag.StringVar(&c.name, "name", "", "")
	flag.StringVar(&c.endpoint, "endpoint", "", "")
	flag.StringVar(&c.caData, "ca-data", "", "")
	flag.StringSliceVar(&c.eksTags, "eks-tags", nil, "Comma-separated KEY=VALUE pairs of EKS control-plane tags")
	flag.StringSliceVar(&c.labelKVs, "labels", nil, "Comma-separated KEY=VALUE pairs of cluster secret labels")

	return &c
}

func InitListClustersFlags(flag *pflag.FlagSet, c *clusterset.ListClustersInput) func() *clusterset.ListClustersInput {
	flag.StringVar(&c.NS, "namespace", c.NS, "")
	flag.StringVar(&c.Selector, "selectoro", "", "")

	return func() *clusterset.ListClustersInput {
		return c
	}
}

func InitCreateTargetGroupBindingFlags(flag *pflag.FlagSet, c *targetgroupbinding.ApplyInput) func() *targetgroupbinding.ApplyInput {
	var labelKVs []string

	flag.BoolVar(&c.DryRun, "dry-run", c.DryRun, "")
	flag.StringVar(&c.ClusterNamespace, "cluster-namespace", c.ClusterNamespace, "")
	flag.StringVar(&c.ClusterName, "cluster-name", c.ClusterName, "")
	flag.StringVar(&c.TargetGroupARN, "target-group-arn", c.TargetGroupARN, "")
	flag.StringVar(&c.Name, "name", c.Name, "")
	flag.StringVar(&c.Namespace, "namespace", c.Namespace, "")
	flag.StringSliceVar(&labelKVs, "labels", nil, "Comma-separated KEY=VALUE pairs of cluster secret labels")

	return func() *targetgroupbinding.ApplyInput {
		labels := map[string]string{}
		for _, kv := range labelKVs {
			split := strings.Split(kv, "=")
			labels[split[0]] = split[1]
		}

		c.Labels = labels
		return c
	}
}

func InitCreateClusterFlags(flag *pflag.FlagSet, c *clusterset.CreateClusterInput) func() *clusterset.CreateClusterInput {
	var labelKVs []string

	flag.BoolVar(&c.DryRun, "dry-run", c.DryRun, "")
	flag.StringVar(&c.NS, "namespace", c.NS, "")
	flag.StringVar(&c.Name, "name", c.Name, "")
	flag.StringVar(&c.Endpoint, "endpoint", c.Endpoint, "")
	flag.StringVar(&c.CAData, "ca-data", "", "")
	flag.StringSliceVar(&labelKVs, "labels", nil, "Comma-separated KEY=VALUE pairs of cluster secret labels")

	return func() *clusterset.CreateClusterInput {
		labels := map[string]string{}
		for _, kv := range labelKVs {
			split := strings.Split(kv, "=")
			labels[split[0]] = split[1]
		}

		c.Labels = labels
		return c
	}
}

func InitDeleteClusterFlags(flag *pflag.FlagSet, c *clusterset.DeleteClusterInput) func() *clusterset.DeleteClusterInput {
	flag.BoolVar(&c.DryRun, "dry-run", c.DryRun, "")
	flag.StringVar(&c.NS, "namespace", c.NS, "")
	flag.StringVar(&c.Name, "name", c.Name, "")

	return func() *clusterset.DeleteClusterInput {
		return c
	}
}

func InitSyncClusterSetFlags(flag *pflag.FlagSet, c *clusterset.SyncInput) func() *clusterset.SyncInput {
	var (
		eksTags  []string
		labelKVs []string
	)

	flag.BoolVar(&c.DryRun, "dry-run", false, "")
	flag.StringVar(&c.NS, "namespace", "", "")
	flag.StringSliceVar(&eksTags, "eks-tags", nil, "Comma-separated KEY=VALUE pairs of EKS control-plane tags")
	flag.StringSliceVar(&labelKVs, "labels", nil, "Comma-separated KEY=VALUE pairs of cluster secret labels")

	return func() *clusterset.SyncInput {
		tags := map[string]string{}
		for _, kv := range eksTags {
			split := strings.Split(kv, "=")
			tags[split[0]] = split[1]
		}

		c.EKSTags = tags

		labels := map[string]string{}
		for _, kv := range labelKVs {
			split := strings.Split(kv, "=")
			labels[split[0]] = split[1]
		}

		c.Labels = labels

		return c
	}
}

func InitListTargetGroupBindingsFlags(flag *pflag.FlagSet, c *targetgroupbinding.ListInput) func() *targetgroupbinding.ListInput {
	flag.StringVar(&c.ClusterName, "cluster-name", "", "")
	flag.StringVar(&c.NS, "namespace", "", "")

	return func() *targetgroupbinding.ListInput {
		return c
	}
}

func InitSyncAWSTargetGroupSetFlags(flag *pflag.FlagSet, c *awstargetgroupset.SyncInput) func() *awstargetgroupset.SyncInput {
	var (
		bindingSelector string
		labelKVs        []string
	)

	flag.BoolVar(&c.DryRun, "dry-run", false, "")
	flag.StringVar(&c.ClusterName, "cluster-name", "", "ArgoCD Cluster name on which we find TargetGroupBinding")
	flag.StringVar(&c.NS, "namespace", "", "Namespace of the ArgoCD Cluster and the generated AWSTargetGroup resources")
	flag.StringVar(&bindingSelector, "targetgroupbinding-selector", "", "Comma-separated KEY=VALUE pairs of TargetGroupBinding resource labels")
	flag.StringSliceVar(&labelKVs, "labels", nil, "Comma-separated KEY=VALUE pairs of AWSTargetGroup labels")

	return func() *awstargetgroupset.SyncInput {
		c.BindingSelector = bindingSelector

		labels := map[string]string{}
		for _, kv := range labelKVs {
			split := strings.Split(kv, "=")
			labels[split[0]] = split[1]
		}

		c.Labels = labels

		return c
	}
}

func InitListAWSTargetGroupsFlags(flag *pflag.FlagSet, c *awstargetgroupset.ListAWSTargetGroupsInput) func() *awstargetgroupset.ListAWSTargetGroupsInput {
	flag.StringVar(&c.NS, "namespace", "", "Namespace of AWSTargetGroup resources")
	flag.StringVar(&c.Selector, "selector", "", "Label selector for AWSTargetGroup resources")

	return func() *awstargetgroupset.ListAWSTargetGroupsInput {
		return c
	}
}

func InitListLatestAWSTargetGroupsFlags(flag *pflag.FlagSet, c *awstargetgroupset.ListLatestAWSTargetGroupsInput) func() *awstargetgroupset.ListLatestAWSTargetGroupsInput {
	flag.StringVar(&c.NS, "namespace", "", "Namespace of AWSTargetGroup resources")
	flag.StringVar(&c.Selector, "selector", "", "Label selector for AWSTargetGroup resources")
	flag.StringVar(&c.Version, "version", "", "Version number without the v prefix of the AWSTargetGroup resources. If omitted, it will fetch all the AWSTargetGroup resources and use the latest version found")
	flag.StringSliceVar(&c.SemverLabelKeys, "semver-label-key", []string{okrav1alpha1.DefaultVersionLabelKey}, "The key of the label as a container of the version number of the group")

	return func() *awstargetgroupset.ListLatestAWSTargetGroupsInput {
		return c
	}
}

func InitRunAnalysisFlags(flag *pflag.FlagSet, c *analysis.RunInput) func() *analysis.RunInput {
	flag.StringVar(&c.AnalysisTemplateName, "template-name", "", "")
	flag.StringVar(&c.NS, "namespace", "", "")
	flag.StringToStringVar(&c.AnalysisArgs, "args", map[string]string{}, "")
	flag.StringToStringVar(&c.AnalysisArgsFromSecrets, "args-from-secrets", map[string]string{}, "A list of secret refs like \"arg-name=secret-name.field-name\" concatenated by \",\"s")

	return func() *analysis.RunInput {
		return c
	}
}

func Run() error {
	cmd := &cobra.Command{
		Use: ApplicationName,
	}

	var listClustersInput func() *clusterset.ListClustersInput
	listClusters := &cobra.Command{
		Use:           "list-clusters",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			clusters, err := clusterset.ListClusters(*listClustersInput())
			if err != nil {
				if !errors.Is(err, okraerror.Error{}) {
					cmd.SilenceUsage = true
				}

				return err
			}

			for _, c := range clusters {
				fmt.Fprintf(os.Stdout, "%v\n", c.Name)
			}

			return nil
		},
	}
	listClustersInput = InitListClustersFlags(listClusters.Flags(), &clusterset.ListClustersInput{})
	cmd.AddCommand(listClusters)

	var createClusterInput func() *clusterset.CreateClusterInput
	createCluster := &cobra.Command{
		Use:           "create-cluster [--namespace ns] [--name name] [--labels k1=v1,k2=v2] [--endpoint https://...] [--ca-data cadata]",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := clusterset.CreateCluster(*createClusterInput()); err != nil {
				if !errors.Is(err, okraerror.Error{}) {
					cmd.SilenceUsage = true
				}

				return err
			}

			return nil
		},
	}
	createClusterInput = InitCreateClusterFlags(createCluster.Flags(), &clusterset.CreateClusterInput{})
	cmd.AddCommand(createCluster)

	var deleteClusterInput func() *clusterset.DeleteClusterInput
	deleteCluster := &cobra.Command{
		Use: "delete-cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterset.DeleteCluster(*deleteClusterInput())
		},
	}
	deleteClusterInput = InitDeleteClusterFlags(deleteCluster.Flags(), &clusterset.DeleteClusterInput{})
	cmd.AddCommand(deleteCluster)

	var createMissingClustersInput func() *clusterset.SyncInput
	createMissingClusters := &cobra.Command{
		Use: "create-missing-clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterset.CreateMissingClusters(*createMissingClustersInput())
		},
	}
	createMissingClustersInput = InitSyncClusterSetFlags(createMissingClusters.Flags(), &clusterset.SyncInput{})
	cmd.AddCommand(createMissingClusters)

	var deleteOutdatedClustersInput func() *clusterset.SyncInput
	deleteOutdatedClusters := &cobra.Command{
		Use: "delete-outdated-clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterset.DeleteOutdatedClusters(*deleteOutdatedClustersInput())
		},
	}
	deleteOutdatedClustersInput = InitSyncClusterSetFlags(deleteOutdatedClusters.Flags(), &clusterset.SyncInput{})
	cmd.AddCommand(deleteOutdatedClusters)

	var syncClusterSetInput func() *clusterset.SyncInput
	syncClusterSet := &cobra.Command{
		Use: "sync-clusterset",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterset.Sync(*syncClusterSetInput())
		},
	}
	syncClusterSetInput = InitSyncClusterSetFlags(syncClusterSet.Flags(), &clusterset.SyncInput{})
	cmd.AddCommand(syncClusterSet)

	var listTargetGroupBindingInput func() *targetgroupbinding.ListInput
	listTargetGroupBindings := &cobra.Command{
		Use: "list-targetgroupbindings",
		RunE: func(cmd *cobra.Command, args []string) error {
			bindings, err := targetgroupbinding.List(*listTargetGroupBindingInput())

			for _, b := range bindings {
				fmt.Fprintf(os.Stdout, "%+v\n", b)
			}

			return err
		},
	}
	listTargetGroupBindingInput = InitListTargetGroupBindingsFlags(listTargetGroupBindings.Flags(), &targetgroupbinding.ListInput{})
	cmd.AddCommand(listTargetGroupBindings)

	var applyTargetGroupBindingInput func() *targetgroupbinding.ApplyInput
	applyTargetGroupBinding := &cobra.Command{
		Use: "apply-targetgroupbinding",
		RunE: func(cmd *cobra.Command, args []string) error {
			binding, err := targetgroupbinding.Apply(*applyTargetGroupBindingInput())

			if binding != nil {
				fmt.Fprintf(os.Stdout, "%+v\n", binding)
			}

			return err
		},
	}
	applyTargetGroupBindingInput = InitCreateTargetGroupBindingFlags(applyTargetGroupBinding.Flags(), &targetgroupbinding.ApplyInput{})
	cmd.AddCommand(applyTargetGroupBinding)

	var createMissingAWSTargetGroupsInput func() *awstargetgroupset.SyncInput
	createMissingAWSTargetGroups := &cobra.Command{
		Use: "create-missing-awstargetgroups",
		RunE: func(cmd *cobra.Command, args []string) error {
			bindings, err := awstargetgroupset.CreateMissingAWSTargetGroups(*createMissingAWSTargetGroupsInput())

			for _, b := range bindings {
				fmt.Fprintf(os.Stdout, "%+v\n", b)
			}

			return err
		},
	}
	createMissingAWSTargetGroupsInput = InitSyncAWSTargetGroupSetFlags(createMissingAWSTargetGroups.Flags(), &awstargetgroupset.SyncInput{})
	cmd.AddCommand(createMissingAWSTargetGroups)

	var deleteOutdatedAWSTargetGroupsInput func() *awstargetgroupset.SyncInput
	deleteOutdatedAWSTargetGroups := &cobra.Command{
		Use: "delete-outdated-awstargetgroups",
		RunE: func(cmd *cobra.Command, args []string) error {
			deleted, err := awstargetgroupset.DeleteOutdatedAWSTargetGroups(*deleteOutdatedAWSTargetGroupsInput())

			for _, b := range deleted {
				fmt.Fprintf(os.Stdout, "%+v\n", b)
			}

			return err
		},
	}
	deleteOutdatedAWSTargetGroupsInput = InitSyncAWSTargetGroupSetFlags(deleteOutdatedAWSTargetGroups.Flags(), &awstargetgroupset.SyncInput{})
	cmd.AddCommand(deleteOutdatedAWSTargetGroups)

	var syncAWSTargetGroupSetInput func() *awstargetgroupset.SyncInput
	syncAWSTargetGroupSet := &cobra.Command{
		Use: "sync-awstargetgroupset",
		RunE: func(cmd *cobra.Command, args []string) error {
			bindings, err := awstargetgroupset.Sync(*syncAWSTargetGroupSetInput())

			for _, b := range bindings {
				fmt.Fprintf(os.Stdout, "%+v\n", b)
			}

			return err
		},
	}
	syncAWSTargetGroupSetInput = InitSyncAWSTargetGroupSetFlags(syncAWSTargetGroupSet.Flags(), &awstargetgroupset.SyncInput{})
	cmd.AddCommand(syncAWSTargetGroupSet)

	var listTargetGroupsInput func() *awstargetgroupset.ListAWSTargetGroupsInput
	listTargetGroups := &cobra.Command{
		Use: "list-awstargetgroups",
		RunE: func(cmd *cobra.Command, args []string) error {
			bindings, err := awstargetgroupset.ListAWSTargetGroups(*listTargetGroupsInput())

			for _, b := range bindings {
				fmt.Fprintf(os.Stdout, "%+v\n", b)
			}

			return err
		},
	}
	listTargetGroupsInput = InitListAWSTargetGroupsFlags(listTargetGroups.Flags(), &awstargetgroupset.ListAWSTargetGroupsInput{})
	cmd.AddCommand(listTargetGroups)

	var listLatestTargetGroupsInput func() *awstargetgroupset.ListLatestAWSTargetGroupsInput
	listLatestTargetGroups := &cobra.Command{
		Use: "list-latest-awstargetgroups",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, bindings, err := awstargetgroupset.ListLatestAWSTargetGroups(*listLatestTargetGroupsInput())

			for _, b := range bindings {
				fmt.Fprintf(os.Stdout, "%+v\n", b)
			}

			return err
		},
	}
	listLatestTargetGroupsInput = InitListLatestAWSTargetGroupsFlags(listLatestTargetGroups.Flags(), &awstargetgroupset.ListLatestAWSTargetGroupsInput{})
	cmd.AddCommand(listLatestTargetGroups)

	cmd.AddCommand(syncAWSApplicationLoadBalancerConfigCommand())
	cmd.AddCommand(syncCellCommand())
	cmd.AddCommand(syncPauseCommand())
	cmd.AddCommand(cancelPauseCommand())
	cmd.AddCommand(updateAnalysisRunCommand())
	cmd.AddCommand(createOrUpdateCellCommand())

	var runAnalysisInput func() *analysis.RunInput
	runAnalysis := &cobra.Command{
		Use: "run-analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			run, err := analysis.Run(*runAnalysisInput())

			if run != nil {
				fmt.Fprintf(os.Stdout, "%+v\n", *run)
			}

			return err
		},
	}
	runAnalysisInput = InitRunAnalysisFlags(runAnalysis.Flags(), &analysis.RunInput{})
	cmd.AddCommand(runAnalysis)

	m := &manager.Manager{}

	controllerManager := &cobra.Command{
		Use: "controller-manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := m.Run()
			if !errors.Is(err, okraerror.Error{}) {
				cmd.SilenceUsage = true
			}
			return err
		},
	}
	m.AddPFlags(controllerManager.Flags())
	cmd.AddCommand(controllerManager)

	err := cmd.Execute()

	return err
}
