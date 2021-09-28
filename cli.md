# The Okra CLI

`okra` is a single-binary executable that provides a CLI for `Okra`.

It is currently used to run and test various operations used by various okra controllers, providing following commands.

- [create argocdclustersecret](#create-argocdclustersecret)
- [create awstargetgroup](#create-awstargetgroup)
- [create cell](#create-cell)
- [sync cell](#sync-cell)

## create argocdclustersecret

`create argocdclustersecret` command replicates the behaviour of `clusterset` controller.

When `--dry-run` is provided, it emits a Kubernetes manifest YAML of a Kubernetes secret to stdout so that it can be applied using e.g. `kubectl apply -f -`.

### create argocdclustersecret --awseks-cluster-name $NAME --version $VERSION

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

### create argocdclustersecret --awseks-cluster-tags $KEY=$VALUE --version-from-tag $VERSION_TAG_KEY

This command calls AWS EKS ListClusters API to list all the clusters in your AWS account, and then calls `DescribeCluster` API for each cluster to get the tags of the respective cluster. For each cluster whose tags matches the selector specified via `--cluster-tags`, the command command creates a ArgoCD cluster secret whose name is set to the same value as the name of the EKS cluster.

The `okra.mumo.co/version` label value of the resulting cluster secret is generated from the value of the tag whose key matches `$VERSION_TAG_KEY`.

## create awstargetgroup

`create argocdclustersecret` command replicates the behaviour of `awstargetgroup` controller.

When `--dry-run` is provided, it emits a Kubernetes manifest YAML of a `AWSTargetGroup` resource to stdout so that it can be applied using e.g. `kubectl apply -f -`.

### create awstargetgroup $$RESOURCE_NAME --arn $ARN --label role=$ROLE

This command creates a `AWSTargetGroup` resource whose name is `$RESOURCE_NAME` and the target group arn is `$ARN` and the `role` label is set to `$ROLE`.

### create awstargetgroup $RESOURCE_NAME --cluster-name $NAME --arn-from-target-group-binding-name $TG_BINDING_NAME --label role=$ROLE

This command get the `TargetGroupBinding` resource named `$TG_BINDING_NAME` from the targeted EKS cluster, and create a `AWSTargetGroup` resource with the target group ARN found in the binding resource.

## create cell

`create cell` is a convenient command to generate and deploy a `Cell` resource.

### create cell $NAME --awsalb-listener-arn $LISTENER_ARN --awstargetgroup-selector role=web --replicas $REPLICAS

## sync cell

`sync cell` replicates the behaviour of `cell` controller.

### sync cell $NAME

This command loads `Cell` resource named $NAME`.

It then fetches all the `AWSTargetGroup` resources that matches the selector (`role=web` for example), group the matched resources by `okra.mumo.co/version` tag values (by default), and sort the groups in an descending order of the semver assuming the tag value contains a semver.

When the group with the newest version has `$REPLICAS` or more target groups in it, it starts updating the AWS ALB listener denoted by `$LISTENER_ARN`.

Before actually updating the listener, it runs analysis. A listener update is pended until there are enough number of successful Analysis runs that happened after the lastest ALB forward config update. Therefore, to complete the canary deployment, you usually need to run `sync cell` several or dozen times.

`sync cell` uses the custom resource status as a state store. More concretely, the last ALB forward config update time and the history of analysis runs and the results are stored in the `Cell` resource's status.
