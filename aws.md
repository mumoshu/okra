# Okra-provided Custom Resources

`okra` provides the following Kubernetes CRDs:

- [Cell](#cell)
  - [Cell with AWSApplicationLoadBalancer](#cell-with-awsapplicationloadbalancer)
  - [Cell with AWSNetworkLoadBalancer](#cell-with-awsnetworkloadbalancer)
- [ClusterSet](#clusterset)
- [AWSTargetGroupSet](#awstargetgroupset)
- [AWSTargetGroup](#awstargetgroup)

# Cell

`Cell` represents a cell in a Cell-based architecture. Each cell is assumed to consist of one or more Kubernetes clusters that is called a cluster "replica".

You usually configure ArgoCD and its ApplicationSet resource so that it can automatically discover and deploy onto Kubernetes clusters newly created by any tool like Terraform, AWS CDK, Pulumi, and so on.

`cell-controller` comes next. It detects N clusters with the latest version number where N is denoted by `cell.spec.replicas`. It then starts updating some loadbalancer configuration while running various analysis on the application running on the new clusters.

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
  # replicas: N
  # versionedBy:
  #   label: version
  #   creationTimestamp: {}
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

Unlike it's Application counterpart, this resource has support for BlueGreen strategy only due to the limitation of Network Load Balancer.

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
    # versionedBy:
    #   label: version
    #   creationTimestamp: {}
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
  - eks:
      tags:
        role: "web"
  template:
    metadata:
      labels:
        role: "{{role}}"
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
  - eks:
      tags:
        role: "web"
  template:
    metadata:
      name: web-"{{.eks.clusterName}}"
      labels:
        role: "{{.eks.tags.role}}"
    spec:
        targets:
        - name: specificnodeincluster
          clusterName: "{{.eks.clusterName}}"
          nodeSelector:
            type: node
          port: 8080
```

# AWSTargetGroup

`AWSTargetGroup` represents a desired state of an existing or dynamically generated AWS target group.

The controller reconciles an `AWSTargetGroup` resource by discovering the target clusters, then discovers pods and nodes in the clusters as targets, and finally creates or updates a target group to register the discovered targets.

Usually, this resource is managed by `AWSTargetGroupSet`. One that is managed by `AWSTargetGroupSet` would look like the below:

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
kind: AWSTargetGroup
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

