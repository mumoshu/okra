# Okra-provided Custom Resources

`okra` provides the following Kubernetes CRDs:

- [Cell](#cell)
  - [Cell with AWSApplicationLoadBalancer](#cell-with-awsapplicationloadbalancer)
  - [Cell with AWSNetworkLoadBalancer](#cell-with-awsnetworkloadbalancer)
- [ClusterSet](#clusterset)
- [AWSTargetGroupSet](#awstargetgroupset)
- [AWSTargetGroup](#awstargetgroup)
- [AWSTargetGroupBinding](#awstargetgroup)

# Cell

`Cell` represents a cell in a Cell-based architecture. Each cell is assumed to consist of one or more Kubernetes clusters that is called a cluster "replica".

However, you don't see clusters in a Cell spec. It is generalized to only require target groups, loadbalancer, and metrics related configuration.

You can bring your own provisioning tool to create EKS cluster with or without target groups. If you want to let Okra create target groups for you, use Okra's [ClusterSet](#clusterset).

You usually configure ArgoCD and its ApplicationSet resource so that it can automatically discover and deploy onto Kubernetes clusters newly created by your provisoning toool. Such tools include Terraform, AWS CDK, Pulumi, and so on.

`cell-controller` comes only after the target groups are created. It detects N target groups before rollout. It firstly groups target groups by the value of the label denoted by `cell.spec.versionedBy.label`, and it then sorts groups of target groups by the version number in the label. Once there are N target groups for the latest version number, where N is denoted by `cell.spec.replicas`, it starts updating the loadbalancer configuration. It concurrently runs various analysis on the application running (behind the target groups|on the new clusters), to ensure safe rollout.

## Cell with AWSApplicationLoadBalancer

`AWSApplicationLoadBalancerTargetDeployment` represents a set of AWS target groups that is routed via an existing AWS Application Load Balancer.

The controller reconciles this resource to discover a the latest set of target groups to where the traffic is routed by the Application Load Balancer.

The only supported `updateStrategy` is `Canary`, which gradually migrates traffic while running analysis.

```yaml
apiVersion: okra.mumoshu.github.io/v1alpha1
kind: Cell
metadata:
  name: web
spec:
  ingress:
    type: AWSApplicationLoadBalancer
    awsApplicationLoadBalancer:
      listenerARN: ...
      targetGroupSelector:
        matchLabels:
          role: web
  # Okra by default starts a rollout process once it finds one target group with newer `okra.mumo.co/version` label value
  # Specify 2 or more to delay the rollout until that number of new target groups are available.
  # replicas: 2
  #
  # Specify the exact version number to rollback, or stick with non-latest version
  # version: 1.2.3
  updateStrategy:
    type: Canary
    # Canary uses the set of target groups whose labels contains
    # `selector.MatchLabels` and the size of the set is equal or greater than N.
    # When there are two or more such sets exist, the one that has the largest version is used.
    canary:
      # See https://argoproj.github.io/argo-rollouts/features/analysis/#background-analysis
      analysis:
        templates:
        - templateName: success-rate
        startingStep: 2 # delay starting analysis run until setWeight: 20%
        args:
        - name: service-name
          value: guestbook-svc.default.svc.cluster.local
      steps:
      - stepWeight: 10
      - pause: {duration: 10m}
      - stepWeight: 20
      - analysis:
          templates:
          - templateName: success-rate
          args:
          - name: service-name
            value: guestbook-svc.default.svc.cluster.local
```

`cell-controller` uses `targetGroupSelector` that is available in the cell spec to determine the ARN of the target group for each cluster, and updates the ALB with new forward config.

`cell-controller` gradually updates forward config target group weights, by `stepWeight` on each interval, so that the gradual update happens. Under the hood, it just calls AWS APIs to update ALB Listener Rules.

`AWSApplicationLoadBalancer`'s `status` sub-resource contains all the fields of the `spec` that applied to AWS. `cell-controller` compares `AWSApplicationLoadBalancer.spec` and `AWSApplicationLoadBalancer.status` and move the process forward only after the two becomes in-sync. Otherwise, it might fail to update weights by `stepWeight` when in a temporary AWS failure.

## Cell with AWSNetworkLoadBalancer

`Cell` with `AWSNetworkLoadBalancer` represents the latest AWS target group that is exposed to the client with an AWS Network Load Balancer.

Unlike it's Application counterpart, this resource has support for BlueGreen strategy only, due to the limitation of Network Load Balancer.

```yaml
apiVersion: okra.mumoshu.github.io/v1alpha1
kind: Cell
spec:
  ingress:
    type: AWSNetworkLoadBalancer
    awsApplicationLoadBalancer:
      listenerARN: ...
      targetGroupSelector:
        matchLabels:
          role: web
    replicas: 2
  updateStrategy:
    type: BlueGreen
    // Unlike Canary, BlueGreen uses the latest target group that matches the
    // selector. This is due to the 
    blueGreen:
      steps:
      - analysis:
          templates:
          - templateName: success-rate
          args:
          - name: service-name
            value: guestbook-svc.default.svc.cluster.local
      - promote: {}
```

# ClusterSet

`ClusterSet` auto-discovers EKS clusters and generates ArgoCD cluster secrets.

`clusterset-controller` reconciles this resource, by calling AWS EKS GetCluster API, build a Kubernetes client config from it, and then create ArgoCD cluster secrets.

```yaml
apiVersion: okra.mumoshu.github.io/v1alpha1
kind: ClusterSet
metadata:
  name: cart
spec:
  generators:
  - awseks:
      clusterSelector:
        matchTags:
          role: "web"
  template:
    metadata:
      labels:
        role: "{{.awseks.cluster.name}}"
```

Assuming there were two EKS clusters whose tags contained `role=web`, the following two cluster secrets are generated:

```yaml
kind: Secret
metadata:
  name: cart-web1
  labels:
    role: web
    argocd.argoproj.io/secret-type: cluster
  ownerReferences:
  - apiVersion: $API_VERSION
    blockOwnerDeletion: true
    controller: true
    kind: ClusterSet
    name: cart
```

```yaml
kind: Secret
metadata:
  name: cart-web2
  labels:
    role: web
    argocd.argoproj.io/secret-type: cluster
  ownerReferences:
  - apiVersion: $API_VERSION
    blockOwnerDeletion: true
    controller: true
    kind: ClusterSet
    name: cart
```


# AWSApplicationLoadBalancerConfig

`AWSApplicationLoadBalancerConfig` represents a desired configuration of a specific AWS Application Loadbalancer.

```
kind: AWSApplicationLoadBalancerConfig
metadata:
  name: ...
spec:
  listenerARN: $LISTENER_ARN
  forwardTargetGroups:
  - name: prev
    arn: prev1
    weight: 40
  - name: next
    arn: prev2
    weight: 60
```

`cell-controller` is responsible for gradually updating `forwardConfig` depending on `stepWeight`. The `awsapplicationloadbalancerconfig-controller` updates the target ALB as exactly as described in the config.

# AWSTargetGroupSet

`AWSTargetGroupSet` auto-discovers clusters and generates `AWSTargetGroup`.

This resources has support for the `eks` generator that generates a cluster secret per discoverered EKS cluster.

The below example would result in creating one AWSTargetGroup per cluster and each group consists of all the nodes whose label matches `type=node`.

```yaml
apiVersion: okra.mumoshu.github.io/v1alpha1
kind: AWSTargetGroupSet
metadata:
  name: web
spec:
  generators:
  - awseks:
      clusterSelector:
        matchLabels:
          role: "web"
      bindingSelector:
        matchLabels:
          role: "web"
  # template is a template for dynamically generated AWSTargetGroup resources
  template:
    metadata:
      labels:
        role: "{{.awseks.cluster.tags.role}}"
  # bindingTemplate is optional, and used only when you want to dynamically generate AWSTargetGroupBinding
  bindingTemplate:
    metadata:
      name: web-"{{.awseks.cluster.name}}"
      labels:
        role: "{{.awseks.cluster.tags.role}}"
    spec:
        targets:
        - name: specificnodeincluster
          clusterName: "{{.awseks.cluster.name}}"
          nodeSelector:
            type: node
          port: 8080
```

# AWSTargetGroup

`AWSTargetGroup` represents an existing AWS target group that is managed by okra or by an external controller like `aws-load-balancer-controller` or `terraform` and so on.

The purpose of this resource is to let the cell controller to know about the new and the old target groups for canary deployments.

```yaml
apiVersion: okra.mumoshu.github.io/v1alpha1
kind: AWSTargetGroup
metadata:
  name: web-cluster1
  labels:
    role: web
  ownerReferences:
  - apiVersion: $API_VERSION
    blockOwnerDeletion: true
    controller: true
    kind: AWSTargetGroupSet
    name: web
spec:
  arn: $TARGET_GROUP_ARN
```

# AWSTargetGroupBinding

`AWSTargetGroupBinding` represents a desired state of an existing or dynamically generated AWS target group.

The controller reconciles an `AWSTargetGroupBinding` resource by discovering the target clusters, then discovers pods and nodes in the clusters as targets, and finally creates or updates a target group to register the discovered targets.

Usually, this resource is managed by `AWSTargetGroupSet`. One that is managed by `AWSTargetGroupSet` would look like the below:

```yaml
apiVersion: okra.mumoshu.github.io/v1alpha1
kind: AWSTargetGroupBinding
metadata:
  name: web-cluster1
  labels:
    role: web
  ownerReferences:
  - apiVersion: $API_VERSION
    blockOwnerDeletion: true
    controller: true
    kind: AWSTargetGroupSet
    name: web
spec:
  targetGroupARN: $TARGET_GROUP_ARN
  targets:
  - name: specificnodesincluster
    clusterName: $EKS_CLUSTER_NAME_1
    nodeSelector:
      labels:
        type: node
    port: 8080
status:
  targets:
  - name: specificnodesincluster
    ids:
    - $INSTANCE_ID_1
    - $INSTANCE_ID_2
    port: 8080
```

Another use-case of this resource is to let the controller register any targets as you like. This is useful when you're managing a single target group in e.g. Terraform or CloudFormation and you'd like to register all the nodes in multiple clusters to the target group.

```yaml
apiVersion: okra.mumoshu.github.io/v1alpha1
kind: AWSTargetGroupBinding
metadata:
  labels:
    role: web
spec:
  targetGroupARN: $TARGET_GROUP_ARN
  targets:
  - name: specificips
    ids:
    - $IP_ADDRESS
    port: 8080
  - name: specificnodeincluster
    clusterName: $EKS_CLUSTER_NAME_1
    nodeName: $NODE_NAME
    port: 8080
  - name: nodesincluster
    clusterSelector:
      matchLabels:
        role: web
    nodeSelector:
      matchLabels:
        role: web
    port: 8080
status:
  targetGroupARN: $TARGET_GROUP_ARN
  targets:
  - name: specificip
    port: 8080
    ids:
    - $IP_ADDRESS
  - name: specificnodeincluster
    port: 8080
    ids:
    - $CLUSTER_1_NODE_INSTANCE_ID
  - name: nodesincluster
    port: 8080
    clusters:
    - webcluster1
    - webcluster2
    ids:
    - $WEBCLUSTER1_NODE_INSTANCE_1_ID
    - $WEBCLUSTER2_NODE_INSTANCE_1_ID
```

Although this is very similar to [aws-load-balancer-controller 's TargetGroupBinding](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.1/guide/targetgroupbinding/targetgroupbinding/), it's different in that `AWSTargetGroup` does not require an existing target group and it can also build a multi-cluster target group.

# `Check`

```
checks:
- name: dd
  analysis
    query: |
      ... {{.Vars.eksClusterName}}
      ... {{.Vars.albListenerARN}} ...{{.Vars.targetGroupARN}}
```

translates to the below if queries are different across clusters:

```
kind: Check
metadata:
  name: mysystem-dd-mycluster-web-1-v2
  annotations:
    analysis-hash: somehash
spec:
  analysis
    interval: 10s
    query: |
      ... mycluster-web-1-v2
      ... listenerARN1 ... targetGroupARN1
    max: 0.1
status:
  analysis:
    observedHash: somehash
    results:
    - ...
```

or the below if the queries are equivalent across clusters:

```
kind: Check
metadata:
  name: mysystem-dd
  annotations:
    analysis-hash: somehash
spec:
  analysis
    interval: 10s
    query: |
      ... v2 ...
    max: 0.1
status:
  lastPassed: true
  lastRunTime: iso3339datetimestring
  analysis:
    observedHash: somehash
    results:
    - ...
```

`check.status.lastPassed` becemes `true` when and only when the last check passed. `lastRunTime` contains the time when the last check ran.

`status.observedHash=somehash` equals to the value in `analysis-hash: somehash` after sync. You can leverage this to make sure that the last check was run with the latest query.
