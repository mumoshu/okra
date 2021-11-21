package clclient

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	elbv2v1beta1 "github.com/mumoshu/okra/api/elbv2/v1beta1"
	rolloutsv1alpha1 "github.com/mumoshu/okra/api/rollouts/v1alpha1"
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
	"github.com/mumoshu/okra/pkg/okraerror"
	"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = okrav1alpha1.AddToScheme(scheme)
	_ = rolloutsv1alpha1.AddToScheme(scheme)
	_ = elbv2v1beta1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func Scheme() *runtime.Scheme {
	return scheme
}

func New() (client.Client, error) {
	restConfig := config.GetConfigOrDie()
	return NewFromRestConfig(restConfig)
}

func Init(c client.Client, s *runtime.Scheme) (client.Client, *runtime.Scheme, error) {
	if c == nil {
		var err error

		c, err = New()
		if err != nil {
			return nil, nil, err
		}
	}

	if s == nil {
		s = Scheme()
	}

	return c, s, nil
}

func NewFromRestConfig(config *rest.Config) (client.Client, error) {
	cl, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, okraerror.New(fmt.Errorf("creating client from rest config: %w", err))
	}

	return cl, nil
}

func NewFromBytes(kubeconfig []byte) (client.Client, error) {
	clCfg, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, okraerror.New(err)
	}

	clClCfg, err := clCfg.ClientConfig()
	if err != nil {
		return nil, okraerror.New(err)
	}

	cl, err := client.New(clClCfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, okraerror.New(err)
	}

	return cl, nil
}

// NewFromClusterSecret returns a controller-runtime client that is able to interact with
// the Kubernetes API server via a dynamic interface.
// If the cluster access ends up with `Unauthorized` errors, try isolating the cause by running
// kubectl with the connection details, like
//   kubectl --token k8s-aws-v1.REDACTED --server https://REDACTED.REDACTED.ap-northeast-1.eks.amazonaws.com get no
// where the token is obtained by
//   aws eks get-token --cluster-name $CLUSTER_NAME
// and server
//   aws eks describe-cluster --name $CLUSTER_NAME
// CDK users- Beware that you need to use the CDK role when accessing the cluster
//   https://github.com/aws/aws-cdk/issues/3752#issuecomment-525213763
// In that case `aws eks get-token` command should include `--role-arn` flag like
//   aws eks get-token --cluster-name $CLUSTER_NAME --role-arn $ROLE_ARN
// If the get-token command ends up failing with `AccessDenied`, you will need to recreate the cluster with a proper `masterRole` config.
// See below for more about that.
//   https://docs.aws.amazon.com/cdk/api/latest/docs/aws-eks-readme.html#masters-role
func NewFromClusterSecret(clusterSecret corev1.Secret) (client.Client, error) {
	cluster, err := SecretToCluster(&clusterSecret)
	if err != nil {
		return nil, fmt.Errorf("secret to cluster: %w", err)
	}

	if cluster.Config.AWSAuthConfig != nil && cluster.Config.AWSAuthConfig.ClusterName != "" {
		if _, err := exec.LookPath("aws"); err != nil {
			return nil, okraerror.New(fmt.Errorf("looking for executable \"aws\": %v", err))
		}
	}

	return NewFromRestConfig(cluster.RESTConfig())
}

// SecretToCluster converts a secret into a Cluster object
// Derived from https://github.com/argoproj/argo-cd/blob/2147ed3aea727ba128df629d53a1d25fd0f6927c/util/db/cluster.go#L290
func SecretToCluster(s *corev1.Secret) (*Cluster, error) {
	const (
		// AnnotationKeyRefresh is the annotation key which indicates that app needs to be refreshed. Removed by application controller after app is refreshed.
		// Might take values 'normal'/'hard'. Value 'hard' means manifes
		// Copied from https://github.com/argoproj/argo-cd/blob/cc4eea0d6951f1025c9ebb487374658186fa8984/pkg/apis/application/v1alpha1/application_annotations.go#L4-L6
		AnnotationKeyRefresh string = "argocd.argoproj.io/refresh"

		// LabelKeySecretType contains the type of argocd secret (currently: 'cluster', 'repository', 'repo-config' or 'repo-creds')
		// Copied from https://github.com/argoproj/argo-cd/blob/3c874ae065c14102003d041d76d4a337abd72f1e/common/common.go#L107-L108
		LabelKeySecretType = "argocd.argoproj.io/secret-type"

		// AnnotationKeyManagedBy is annotation name which indicates that k8s resource is managed by an application.
		// Copied from https://github.com/argoproj/argo-cd/blob/3c874ae065c14102003d041d76d4a337abd72f1e/common/common.go#L122-L123
		AnnotationKeyManagedBy = "managed-by"
	)

	var config ClusterConfig
	if len(s.Data["config"]) > 0 {
		err := json.Unmarshal(s.Data["config"], &config)
		if err != nil {
			return nil, err
		}
	}

	var namespaces []string
	for _, ns := range strings.Split(string(s.Data["namespaces"]), ",") {
		if ns = strings.TrimSpace(ns); ns != "" {
			namespaces = append(namespaces, ns)
		}
	}
	var refreshRequestedAt *metav1.Time
	if v, found := s.Annotations[AnnotationKeyRefresh]; found {
		requestedAt, err := time.Parse(time.RFC3339, v)
		if err != nil {
			log.Warnf("Error while parsing date in cluster secret '%s': %v", s.Name, err)
		} else {
			refreshRequestedAt = &metav1.Time{Time: requestedAt}
		}
	}
	var shard *int64
	if shardStr := s.Data["shard"]; shardStr != nil {
		if val, err := strconv.Atoi(string(shardStr)); err != nil {
			log.Warnf("Error while parsing shard in cluster secret '%s': %v", s.Name, err)
		} else {
			shard = pointer.Int64Ptr(int64(val))
		}
	}

	// copy labels and annotations excluding system ones
	labels := map[string]string{}
	if s.Labels != nil {
		for k, v := range s.Labels {
			labels[k] = v
		}
		delete(labels, LabelKeySecretType)
	}
	annotations := map[string]string{}
	if s.Annotations != nil {
		for k, v := range s.Annotations {
			annotations[k] = v
		}
		delete(annotations, AnnotationKeyManagedBy)
	}

	cluster := Cluster{
		ID:                 string(s.UID),
		Server:             strings.TrimRight(string(s.Data["server"]), "/"),
		Name:               string(s.Data["name"]),
		Namespaces:         namespaces,
		ClusterResources:   string(s.Data["clusterResources"]) == "true",
		Config:             config,
		RefreshRequestedAt: refreshRequestedAt,
		Shard:              shard,
		Project:            string(s.Data["project"]),
		Labels:             labels,
		Annotations:        annotations,
	}
	return &cluster, nil
}
