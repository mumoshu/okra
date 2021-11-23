# The Okra CLI

`okra` is a single-binary executable that provides a CLI for `Okra`.

It is currently used to run and test various operations used by various okra controllers, providing following commands.

- [create cluster](#create-cluster)
- [create awstargetgroup](#create-awstargetgroup)
- [list targetgroupbindings](#list-targetgroupbindings)
- [list awstargetgroups](#list-awstargetgroups)
- [list latest-awstargetgroups](#list-awslatest-targetgroups)
- [create cell](#create-cell)
- [sync cell](#sync-cell)
- [create awsapplicationloadbalancerconfig](#create-awsapplicationloadbalancerconfig)
- [sync awsapplicationloadbalancerconfig](#sync-awsapplicationloadbalancerconfig)
- [run analysis](#run-analysis)
- [create analysis](#create-analysis)
- [sync analysis](#sync-analysis)

## controller-manager

This command runs okra's controller-manager that is composed of several Kubernetes controllers that powers [CRDs](#/crd.md).

## create cluster

`create cluster` command replicates the behaviour of `clusterset` controller.

When `--dry-run` is provided, it emits a Kubernetes manifest YAML of a Kubernetes secret to stdout so that it can be applied using e.g. `kubectl apply -f -`.

### create cluster --awseks-cluster-name $NAME --version $VERSION

This command calls AWS EKS DescribeCluster API on the EKS cluster whose name equals to `$NAME` and use the CA data of the cluster to create an ArgoCD cluster secret. The cluster secret is named `$NAME`. ArgoCD ApplicationSet can discover the cluster secret and start a deployment.

`$VERSION` is a semantic version of the cluster like `1.0.0`. It's used by the cell controller to group clusters by the version number.

This command create a Kubernetes secret named `$NAME`, labeled with `argocd.argoproj.io/secret-type: "cluster"`, whose `data` looks like the below:

```yaml
kind: Secret
metadata:
  name: $NAME
  labels:
    argocd.argoproj.io/secret-type: "cluster"
    okra.mumo.co/version: "1.0.0"
stringData:
  name: $ARGOCD_CLUSTER_NAME
  server: $SERVER
  confi: |
    {
      "awsAuthConfig": {
        "clusterName": "$EKS_CLUSTER_NAME"
      },
      "tlsClientConfig": {
        "insecure": false,
        "caData": "$BASE64_ENCODED_CA_DATA"
      }
    }
```

When `--add-target-group-annotations` is provided, the resulting cluster secret can be annotated with `okra.mumoshu.github.io/target-group/NAME: {"target-group-arn":"TG_ARN"}`, where `NAME` can be any id that can be a part of the annotation key, and `TG_ARN` is the ARN of the target group associated to the cluster.

### create cluster --awseks-cluster-tags $KEY=$VALUE --version-from-tag $VERSION_TAG_KEY

This command calls AWS EKS ListClusters API to list all the clusters in your AWS account, and then calls `DescribeCluster` API for each cluster to get the tags of the respective cluster. For each cluster whose tags matches the selector specified via `--cluster-tags`, the command command creates a ArgoCD cluster secret whose name is set to the same value as the name of the EKS cluster.

The `okra.mumo.co/version` label value of the resulting cluster secret is generated from the value of the tag whose key matches `$VERSION_TAG_KEY`.

## create awstargetgroup

`create argocdclustersecret` command replicates the behaviour of `awstargetgroup` controller.

When `--dry-run` is provided, it emits a Kubernetes manifest YAML of a `AWSTargetGroup` resource to stdout so that it can be applied using e.g. `kubectl apply -f -`.

### create awstargetgroup $$RESOURCE_NAME --arn $ARN --labels role=$ROLE

This command creates a `AWSTargetGroup` resource whose name is `$RESOURCE_NAME` and the target group arn is `$ARN` and the `role` label is set to `$ROLE`.

### create-missing-awstargetgroups --cluster-name $NAME --target-group-binding-selector name=$TG_BINDING_NAME --labels role=$ROLE

This command get the `TargetGroupBinding` resource named `$TG_BINDING_NAME` from the targeted EKS cluster, and create a `AWSTargetGroup` resource with the target group ARN found in the binding resource. Each `AWSTargetGroup` gets the `role=$ROLE` label as specified by the `--labels` flag.

The part that finds `TargetGroupBinding` can be run independently with [list targetgroupbindings](#list-targetgroupbindings).

### create-outdated-awstargetgroups --cluster-name $NAME --target-group-binding-selector name=$TG_BINDING_NAME --labels role=$ROLE

This command gets the `TargetGroupBinding` resource named `$TG_BINDING_NAME` from the targeted EKS cluster, list `AWSTargetGroup` resources whose labels contains `role=$ROLE` on the management cluster, and delete any `AWSTargetGroup` resources that doesn't have corresponding `TargetGroupBinding`.

The part that finds `TargetGroupBinding` can be run independently with [list targetgroupbindings](#list-targetgroupbindings).

## list targetgroupbindings

This command fetches and outputs all the `TargetGroupBinding` resources in the target cluster. The target cluster is denoted by the name of an ArgoCD cluster secret.

## list awstargetgroups

This command fetches and outputs all the `AWSTargetGroup` resources in the management cluster.

## list latest-awstargetgroups

This command outputs latest `AWSTargetGroup` resources in the management cluster.

It does so by firstly fetching all the `AWSTargetGroup` resources that matches the selector (`role=web` for example), group the matched resources by `okra.mumo.co/version` tag values (by default), and sort the groups in an descending order of the semver assuming the tag value contains a semver.

## create cell

`create cell` is a convenient command to generate and deploy a `Cell` resource.

### create cell $NAME --awsalb-listener-arn $LISTENER_ARN --awstargetgroup-selector role=web --replicas $REPLICAS

## sync cell

`sync cell` runs the main reconcilation logic of `cell-controller`.

### sync cell --name $NAME --target-group-selector $SELECTOR

This command syncs `Cell` resource named `$NAME` with various settings.

It starts by fetching all the `AWSTargetGroup` resources that matches the selector, group the matched resources by `okra.mumo.co/version` tag values (by default), and sort the groups in an descending order of the semver assuming the tag value contains a semver. This part can be run independently with [list latest-targetgroupbindings](#list-latest-awstargetgroups).

For example, if the selector was `role=web`, it will fetch all the `AWSTargetGroup` resources whose `metadata.labels` matches the selector. It then groups up the groups by the `okra.mumo.co/version` tag values. Say the version tag values are `v1.0.0` and `v1.1.0`, the group of the newest version `v1.1.0` comes first hence becomes the next deployment candidate.

When the group with the newest version has the desired number of target groups denoted by `spec.replicas`, it starts updating the AWS ALB listener denoted by `$LISTENER_ARN`.

Before updating the listener, it firstly ensures that there's exactly one loadbalancer config resource. If it didn't find one, it creates one, which is either an `AWSApplicationLoadBalancerConfig` or `AWSNetworkLoadBalancerConfig` resource depending on the loadbalancer specified in the `Cell` resource.

The initial config's spec field is derived from the current state of the loadbalancer obtained by calling AWS API. The creation part can be run independently by using [create awsapplicationloadbalancerconfig](#create-awsapplicationloadbalancerconfig).

If there was an `AWSApplicationLoadBalancerConfig` resource whose `status.phase` is still `Updating`, the command exists with code 0 without creating another `AWSApplicationLoadBalancerConfig` resource. To complete the config update, run [sync awsapplicationloadbalancerconfig](#sync-awsapplicationloadbalancerconfig).

If there's an `AWSApplicationLoadBalancerConfig` resource and its `status.phase` is `Updated` or `Created`, an update starts. An update works differently depending on the current step index. The current step index is either derived from `cell.status` or the `--step-index` flag of this command.

If the current step has `stepWeight`, it updates target groups' weights. The desired target groups' weights are computed from the step index. The controller sums up all the `stepWeight` values of the steps from 0 to the current index for that.

If and only if the desired weights are different from the current weights, it commits a listener update.

More concretely, when the listener is being updated, it compares the current target group ARNs and weights stored in `AWSApplicationLoadBalancerConfig` against the desired target group ARNs and their weights computed by the `cell-controller`, determining the next state with the updated target group ARNs and weights. When in the controller, this happens on each reconcilation loop, so that the weights looks like changing gradually.

If the current step has `sleep`, it exists after updating `cell.status`.

If the current step has `analysis`, it creates a new analysis run from it and exits.

In any case, if previous step has `sleep`, it loads the start and end time of the sleep from `cell.status`, and exits if the current time is before the sleep end time. Similarly, if previous step has `analysis`, it loads the previous analysis run name from `cell.status` and check if the analysis run has phase `Completed`. If it doesn't, it exists.

In other words, it usually either (1)sleep for a while or (2)runs an analysis before updating the listener. If it was a sleep, the next listener update is pended until it the sleep duration elapses.

If it was an analysis, the listener update is pended until there are enough number of successful analysises that happened after the lastest ALB forward config update. To complete the canary deployment, you need to rerun `sync cell` once again after `run analysis` completed. A analysis run can be trigered via [run analysis](#run-analysis) command.

`sync cell` updates `Cell`'s status to signal other K8s controller or clients. It doesn't use the status as a state store.

## create awsapplicationloadbalancerconfig

This command creates a new `AWSApplicationLoadBalancerConfig` resource. To sync it, use [sync awsapplicationloadbalancerconfig](#sync-awsapplicationloadbalancerconfig).

### create awsapplicationloadbalancerconfig $NAME --listener-arn $LISTENER_ARN

This command creates a `AWSApplicationLoadBalancerConfig` resource whose name is `$NAME`. The target group ARNs and their weights are derived from the current state of the loadbalancer and the listener obtained by calling AWS API.

## sync awsapplicationloadbalancerconfig

### sync awsapplicationloadbalancerconfig $NAME

This command loads `AWSApplicationLoadBalancerConfig` resource named `$NAME` and reconciles it.

`sync awsapplicationloadbalancerconfig` uses `AWSApplicationLoadBalancerConfig`'s status to signal `cell-controller` about the completion of the update.

More concretely, `status.phase` is set to `Created`, `Error`, or `Updated` depending on the situation. It is initially `Created`. If the spec has been changed but the controller failed to update it (i.e. AWS API error), the phase becomes `Error`. If the spec update has been successfully applied to the loadbalancer, the phase becomes `Updated`.

## run analysis

`run analysis` creates a Argo Rollout's `AnalysisRun` resource from a `AnalysisTemplate`, and optionally waits for the run to complete.

See the relevant part of [Argo Rollouts documentation](https://argoproj.github.io/argo-rollouts/features/analysis/) for more information about `AnalysisTemplate` and `AnalysisRun`.

### run analysis $NAME --template-name $TEMPLATE_NAME --args key1=val1,key2=val2

This command creates an `AnalysisRun` resource named `$NAME` from the template denoted by `TEMPLATE_NAME`. The run args are populated via `--args`.

Let say you had a template that looks like:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: success-rate
spec:
  args:
  - name: service-name
  - name: prometheus-port
    value: "9090"
  metrics:
  - name: success-rate
    successCondition: result[0] >= 0.95
    provider:
      prometheus:
        address: "http://prometheus.example.com:{{args.prometheus-port}}"
        query: |
          sum(irate(
            istio_requests_total{reporter="source",destination_service=~"{{args.service-name}}",response_code!~"5.*"}[5m]
          )) /
          sum(irate(
            istio_requests_total{reporter="source",destination_service=~"{{args.service-name}}"}[5m]
          ))
```

`okra run analysis run1 --template-name success-rate --args service-name=foo` will create a run like the below.

```yaml
kind: AnalysisRun
metadata:
  name: run1
spec:
  args:
  - name: service-name
    value: foo
  - name: prometheus-port
    value: "9090"
  metrics:
  - name: success-rate
    successCondition: result[0] >= 0.95
    provider:
      prometheus:
        address: "http://prometheus.example.com:{{args.prometheus-port}}"
        query: |
          sum(irate(
            istio_requests_total{reporter="source",destination_service=~"{{args.service-name}}",response_code!~"5.*"}[5m]
          )) /
          sum(irate(
            istio_requests_total{reporter="source",destination_service=~"{{args.service-name}}"}[5m]
          ))
```

`AnalysisRun`'s spec is mostly equivalent to that of `AnalysisTemplate`'s, except that the `service-name` arg's `value` in the template is updated to `foo`. `foo` is from the `--args service-name=foo` and `9090` is from the default value defined in the template's args field.

When `--wait` is provided, the command waits until the run to complete. It's considered complete when `status.phase` is either `Error` or `Succeeded`. If phase was `Error`, the command prints a summary of the last `status.metricResults[].measurements[]` item, and exists with code 1.

Let's say it failed like:

```
  status:
    message: 'metric "success-rate" assessed Error due to consecutiveErrors (5) >
      consecutiveErrorLimit (4): "Error Message: Post "http://prometheus.example.com:9090/api/v1/query":
      dial tcp: lookup prometheus.example.com on 10.96.0.10:53: no such host"'
    metricResults:
    - consecutiveError: 5
      error: 5
      measurements:
      - *snip*
      - finishedAt: "2021-09-28T08:54:29Z"
        message: 'Post "http://prometheus.example.com:9090/api/v1/query": dial tcp:
          lookup prometheus.example.com on 10.96.0.10:53: no such host'
        phase: Error
        startedAt: "2021-09-28T08:54:29Z"
      message: 'Post "http://prometheus.example.com:9090/api/v1/query": dial tcp:
        lookup prometheus.example.com on 10.96.0.10:53: no such host'
      name: success-rate
      phase: Error
    phase: Error
    startedAt: "2021-09-28T08:53:49Z"
```

It writes the message `Post "http://prometheus.example.com:9090/api/v1/query": dial tcp: lookup prometheus.example.com on 10.96.0.10:53: no such host` to stderr and exsits with 1.
