package awstargetgroupset

import (
	"context"
	"fmt"
	"os"

	"github.com/blang/semver"
	"github.com/mumoshu/okra/api/elbv2/v1beta1"
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/clclient"
	"github.com/mumoshu/okra/pkg/okraerror"
	"golang.org/x/xerrors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type Config struct {
	DryRun   bool
	NS       string
	Name     string
	Endpoint string
	CAData   string
	Labels   map[string]string
}

type ClusterSetConfig struct {
	DryRun  bool
	NS      string
	EKSTags map[string]string
	Labels  map[string]string
}

type CreateTargetGroupInput struct {
	DryRun bool
	NS     string
	Name   string
	ARN    string
	Labels map[string]string
}

type SyncInput struct {
	DryRun          bool
	NS              string
	ClusterName     string
	BindingSelector string
	Labels          map[string]string
}

type DeleteInput struct {
	NS     string
	Name   string
	DryRun bool
}

type Provider struct {
	client.Client
}

func New(cl client.Client) *Provider {
	return &Provider{
		Client: cl,
	}
}

func (p *Provider) CreateTargetGroup(config CreateTargetGroupInput) error {
	ns := config.NS
	name := config.Name
	arn := config.ARN
	dryRun := config.DryRun
	labels := config.Labels

	if name == "" {
		return okraerror.New(fmt.Errorf("name is required"))
	}

	object := &okrav1alpha1.AWSTargetGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: okrav1alpha1.AWSTargetGroupSpec{
			ARN: arn,
		},
	}

	if dryRun {
		text, err := yaml.Marshal(object)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "%s\n", text)

		return nil
	}

	if err := p.Client.Create(context.TODO(), object); err != nil {
		return okraerror.New(err)
	}

	fmt.Printf("AWSTargetGroup %q created successfully\n", name)

	return nil
}

func CreateMissingAWSTargetGroups(config SyncInput) ([]SyncResult, error) {
	ns := config.NS
	dryRun := config.DryRun

	clientset, err := clclient.NewClientSet()
	if err != nil {
		return nil, xerrors.Errorf("creating clientset: %w", err)
	}

	kubeclient := clientset.CoreV1().Secrets(ns)

	secret, err := kubeclient.Get(context.TODO(), config.ClusterName, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("listing cluster secrets: %w", err)
	}

	managementClient, err := clclient.New()
	if err != nil {
		return nil, err
	}

	client, err := clclient.NewFromClusterSecret(*secret)
	if err != nil {
		return nil, err
	}

	var bindings v1beta1.TargetGroupBindingList

	optionalNS := ""

	sel, err := labels.Parse(config.BindingSelector)
	if err != nil {
		return nil, xerrors.Errorf("parsing binding selector: %v", err)
	}

	if err := client.List(context.TODO(), &bindings, &runtimeclient.ListOptions{
		Namespace:     optionalNS,
		LabelSelector: sel,
	}); err != nil {
		return nil, okraerror.New(err)
	}

	var objects []okrav1alpha1.AWSTargetGroup

	for _, b := range bindings.Items {
		labels := map[string]string{}

		for k, v := range b.Labels {
			labels[k] = v
		}

		for k, v := range config.Labels {
			labels[k] = v
		}

		objects = append(objects, okrav1alpha1.AWSTargetGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:   b.Name,
				Labels: labels,
			},
			Spec: okrav1alpha1.AWSTargetGroupSpec{
				ARN: b.Spec.TargetGroupARN,
			},
		})
	}

	for _, object := range objects {
		// Manage resource
		if !dryRun {
			err := managementClient.Create(context.TODO(), &object)
			if err != nil {
				if kerrors.IsAlreadyExists(err) {
					fmt.Printf("AWSTargetGroup %q has no change\n", object.Name)
				} else {
					fmt.Fprintf(os.Stderr, "Failed creating object: %+v\n", object)
					return nil, okraerror.New(err)
				}
			} else {
				fmt.Printf("AWSTargetGroup %q created successfully\n", object.Name)
			}
		} else {
			fmt.Printf("AWSTargetGroup %q created successfully (Dry Run)\n", object.Name)
		}
	}

	var created []SyncResult

	return created, nil
}

func Delete(config DeleteInput) error {
	ns := config.NS
	name := config.Name
	dryRun := config.DryRun

	clientset, err := clclient.NewClientSet()
	if err != nil {
		return xerrors.Errorf("creating clientset: %w", err)
	}

	kubeclient := clientset.CoreV1().Secrets(ns)

	if dryRun {
		fmt.Fprintf(os.Stdout, "Cluster secrer %q deleted successfully (dry run)\n", name)

		return nil
	}

	// Manage resource
	err = kubeclient.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Cluster secert %q deleted successfully\n", name)

	return nil
}

func DeleteOutdatedAWSTargetGroups(config SyncInput) ([]SyncResult, error) {
	ns := config.NS
	dryRun := config.DryRun

	clientset, err := clclient.NewClientSet()
	if err != nil {
		return nil, xerrors.Errorf("creating clientset: %w", err)
	}

	kubeclient := clientset.CoreV1().Secrets(ns)

	secret, err := kubeclient.Get(context.TODO(), config.ClusterName, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("listing cluster secrets: %w", err)
	}

	managementClient, err := clclient.New()
	if err != nil {
		return nil, err
	}

	client, err := clclient.NewFromClusterSecret(*secret)
	if err != nil {
		return nil, err
	}

	var bindings v1beta1.TargetGroupBindingList

	optionalNS := ""

	sel, err := labels.Parse(config.BindingSelector)
	if err != nil {
		return nil, xerrors.Errorf("parsing binding selector: %v", err)
	}

	if err := client.List(context.TODO(), &bindings, &runtimeclient.ListOptions{
		Namespace:     optionalNS,
		LabelSelector: sel,
	}); err != nil {
		return nil, okraerror.New(err)
	}

	var objects []okrav1alpha1.AWSTargetGroup

	for _, b := range bindings.Items {
		labels := map[string]string{}

		for k, v := range b.Labels {
			labels[k] = v
		}

		for k, v := range config.Labels {
			labels[k] = v
		}

		objects = append(objects, okrav1alpha1.AWSTargetGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:   b.Name,
				Labels: labels,
			},
			Spec: okrav1alpha1.AWSTargetGroupSpec{
				ARN: b.Spec.TargetGroupARN,
			},
		})
	}

	desiredTargetGroups := map[string]struct{}{}

	for _, obj := range objects {
		desiredTargetGroups[obj.Name] = struct{}{}
	}

	var current okrav1alpha1.AWSTargetGroupList

	if err := managementClient.List(context.TODO(), &current, &runtimeclient.ListOptions{
		Namespace:     optionalNS,
		LabelSelector: sel,
	}); err != nil {
		return nil, okraerror.New(err)
	}

	var deleted []SyncResult

	for _, item := range current.Items {
		name := item.Name

		if _, desired := desiredTargetGroups[name]; !desired {
			if dryRun {
				fmt.Printf("AWSTargetGroup %q deleted successfully (Dry Run)\n", name)
			} else {
				// Manage resource
				err := kubeclient.Delete(context.TODO(), name, metav1.DeleteOptions{})
				if err != nil {
					return nil, err
				}

				fmt.Printf("AWSTargetGroup %q deleted successfully\n", name)
			}

			deleted = append(deleted, SyncResult{
				Name:   name,
				Action: "Delete",
			})
		}
	}

	return deleted, nil
}

type SyncResult struct {
	Name   string
	Action string
}

func Sync(config SyncInput) ([]SyncResult, error) {
	created, err := CreateMissingAWSTargetGroups(config)
	if err != nil {
		return nil, xerrors.Errorf("creating missing cluster secrets: %w", err)
	}

	deleted, err := DeleteOutdatedAWSTargetGroups(config)
	if err != nil {
		return created, xerrors.Errorf("deleting redundant cluster secrets: %w", err)
	}

	all := append([]SyncResult{}, created...)
	all = append(all, deleted...)

	return all, nil
}

type ListLatestAWSTargetGroupsInput struct {
	ListAWSTargetGroupsInput

	SemverLabelKey string
}

type ListAWSTargetGroupsInput struct {
	NS       string
	Selector string
}

func ListLatestAWSTargetGroups(config ListLatestAWSTargetGroupsInput) ([]okrav1alpha1.AWSTargetGroup, error) {
	groups, err := ListAWSTargetGroups(config.ListAWSTargetGroupsInput)
	if err != nil {
		return nil, err
	}

	type entry struct {
		ver    semver.Version
		groups []okrav1alpha1.AWSTargetGroup
	}

	labelKey := config.SemverLabelKey
	if labelKey == "" {
		return nil, fmt.Errorf("missing semver label key")
	}

	var latestVer *semver.Version

	versionedGroups := map[string]entry{}

	for _, g := range groups {
		g := g

		verStr, ok := g.Labels[labelKey]
		if !ok {
			return nil, fmt.Errorf("no semver label found on group: %v", g)
		}

		ver, err := semver.Parse(verStr)
		if err != nil {
			return nil, err
		}

		if latestVer == nil {
			latestVer = &ver
		} else if latestVer.LT(ver) {
			latestVer = &ver
		}

		e := versionedGroups[ver.String()]

		e.ver = ver
		e.groups = append(e.groups, g)

		versionedGroups[ver.String()] = e
	}

	if latestVer == nil {
		return nil, nil
	}

	latest := versionedGroups[latestVer.String()]

	return latest.groups, nil
}

func ListAWSTargetGroups(config ListAWSTargetGroupsInput) ([]okrav1alpha1.AWSTargetGroup, error) {
	managementClient, err := clclient.New()
	if err != nil {
		return nil, err
	}

	sel, err := labels.Parse(config.Selector)
	if err != nil {
		return nil, err
	}

	var list okrav1alpha1.AWSTargetGroupList

	if err := managementClient.List(context.TODO(), &list, &runtimeclient.ListOptions{
		Namespace:     config.NS,
		LabelSelector: sel,
	}); err != nil {
		return nil, err
	}

	return list.Items, nil
}
