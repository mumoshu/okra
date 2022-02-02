# Okra

> This project is under heavy development and has no tagged releases yet.
>
> But I'd still appreciate it if you could help me by testing it and submitting pull requests,
> so that you can get the first release earlier!
>
> We already have a throughout getting-started guide, a working Helm chart, and a container image published at `mumoshu/okra:canary`. So it shouldn't be that hard to give it a shot.

`Okra` is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/) and a set of [CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) which provide advanced multi-cluster appilcation rollout capabilities, such as canary deployment of clusters.

`okra` eases managing a lot of ephemeral Kubernetes clusters.

If you've been using ephemeral Kubernetes clusters and employed blue-green or canary deployments for zero-downtime cluster updates, you might have suffered from a lot of manual steps required. `okra` is intended to automate all those steps.

In a standard scenario, a system update with `okra` would like the below.

- **You** provision one or more new clusters with cluster tags like `name=web-1-v2, role=web, version=v2`
- **Okra** auto-imports the clusters into **ArgoCD**
- **ArgoCD ApplicationSet** deploys your apps onto the new clusters
- **Okra** updates the loadbalancer configuration to gradually migrate traffic to the new clusters, while running various checks to ensure application availability

ToC:

- [How it works](#how-it-works)
- [Getting Started](#getting-started)
- [Comparison with Flagger and Argo Rollouts](#comparison-with-flagger-and-argo-rollouts)
- [CRDs](#crds)
- [CLI](#cli)
- [Why is it named "okra"?](#why-is-it-named-okra)

## Project Status and Scope

`okra` (currently) integrates with AWS ALB and target groups for traffic management, CloudWatch Metrics and Datadog for canary analysis.

`okra` currently works on AWS only, but the design and the implementation of it is generic enough to be capable of adding more IaaS supports. Any contribution around that is welcomed.

Here's the list of possible additional IaaSes that the original author (@mumoshu) has thought of:

- Cluster API
- GKE

Here's the list of possible additional loadbalancers:

- AWS NLB
- Envoy
- [Istio Ingerss Gateway](https://istio.io/latest/docs/tasks/traffic-management/ingress/ingress-control/)
- [ingress-nginx](https://kubernetes.github.io/ingress-nginx/)

## Concepts

`Okra` manages **cells** for you. A cell can be compared to a few things.

A cell is like a Kubernetes pod of containers. A Kubernetes pod an isolated set of containers, where each container usually runs a single application, and you can have two or more pods for availability and scalability. A Okra [`cell`](/crd.md#cell) is a set of Kubernetes clusters, where each cluster runs your application and you can have two or more clusters behind a loadbalancer for horizontal scalability beyond the limit of a single cluster.

A cell is like a storage array but for Kubernetes clusters. You hot-swap a disk in a storage array while running. Similarly, with `okra` you hot-swap a cluster in a cell while keeping your application up and running.

Okra's `cell-contorller` is responsible for managing the traffic shift across clusters.

You give each [`Cell`](/crd.md#cell) a set of settings to discover AWS target groups and configure loadbalancers, and metrics.

The controller periodically discovers AWS target groups. Once there are enough number of new target groups, it then compares the target groups associated to the loadbalancer. If there's any difference, it starts updating the ALB while checking various metrics for safe rollout.

Okra uses Kubernetes CRDs and custom resources as a state store and uses the standard Kubernetes API to interact with resources.

Okra calls various AWS APIs to create and update AWS target groups and update AWS ALB and NLB forward config for traffic management.

### Comparison with Flagger and Argo Rollouts

Unlike `Argo Rollouts` and `Flagger`, in `Okra` there is no notions of "active" and "preview" services for a blue-green deployment, or "canary" and "stable" services for a canary deployment.

It assumes there's one or more target groups per cell. `cell` basically does a canary deployment, where the old set of target groups is consdidered "stable" and the new set of target groups is considered "canary".

In `Flagger` or `Argo Rollouts`, you need to update its K8s resource to trigger a new rollout. In Okra you don't need to do so. You preconfigure its resource and Okra auto-starts a rollout once it discovers enough number of new target groups.

## How it works

`okra` updates your [`Cell`](/crd.md#cell).

A okra `Cell` is composed of target groups and an AWS loadbalancer, and a set of metrics for canary anlysis.

Each target group is tied to a `cluster`, where a `cluster` is a Kubernetes cluster that runs your container workloads.

An `application` is deployed onto `clusters` by `ArgoCD`. The traffic to the `application` is routed via an [AWS ALB](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/introduction.html) in front of `clusters`.

`okra` acts as an application traffic migrator.

It detects new `target groups`, and live migrate traffic by hot-swaping old target groups serving the affected `applications` with the new target groups, while keepining the `applications` up and running.

## Getting Started

- [Install Okra](#install-okra)
- [Create Load Balancer](#create-load-balancer)
- [Provision Kubernetes Clusters](#provision-kubernetes-clusters)
- [Deploy Applications onto Clusters](#deploy-applications-onto-clusters)
  - [Auto-Deploy with ApplicationSet and ClusterSet](#auto-deploy-with-applicationset-and-clusterset)
- [Register Target Groups](#register-target-groups)
  - [Auto-Register Target Groups with AWSTargetGroupSet](#auto-register-target-groups-with-awstargetgroupset)
- [Create Cell](#create-cell)
- [Create and Rollout New Clusters](#create-and-rollout-new-clusters)
- [Analysises and Experiments](#analysises-and-experiments)

### Install Okra

First, you need to provision a Kubernetes cluster that is running ArgoCD, Argo Rollouts, and ArgoCD ApplicationSet controller.
We call it `management cluster` in the following guide.

To deploy required components onto the management cluster, use the following snippet:

```shell
# 1. Install ArgoCD and ApplicationSet
# https://argocd-applicationset.readthedocs.io/en/stable/Getting-Started/#b-install-applicationset-and-argo-cd-together
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj-labs/applicationset/v0.3.0/manifests/install-with-argo-cd.yaml

# 2. Install Argo Rollouts
# https://argoproj.github.io/argo-rollouts/installation/
kubectl create namespace argo-rollouts
kubectl apply -n argo-rollouts -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml
```

Once your management cluster is up and running, install `okra` on it using Helm or Kustomize.

**Option 1: Helm**:

```
$ helm upgrade --install charts/okra -f values.yaml
```

**Option 2: Kustomize**:

```
$ kustomize build config/manager | kubectl apply -f
```

> You can specify okra's container image tag to anything that is available on https://hub.docker.com/r/mumoshu/okra/tags.
>
> For Helm, you do it like `helm upgrade --install charts/okra --set image.tag=$TAG`.

> Note that you need to provide AWS credentials to `okra` as
it calls various AWS API to list and describe EKS clusters, generate Kubernetes API tokens, and interacting with loadbalancers.
>
> For Helm, the simplest (but not recommended in production) way would be to provide `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`:
>
> `values.yaml`:
> ```yaml
> region: ap-northeast-1
> image:
>   tag: "canary"
> additionalEnv:
> - name: AWS_ACCESS_KEY_ID
>   value: "..."
> - name: AWS_SECRET_ACCESS_KEY
>   value: "..."
> ```
>
> For production environments, you'd better use [IAM roles for service accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html) for security reason.

### Create Load Balancer

Create a loadbalancer in front of all the clusters you're going to manage with Okra.

Currently, only AWS Application LoadBalancer is supported.

You can use Terraform, AWS CDK, Pulumi, AWS Console, AWS CLI, or whatever tool to create the loadbalancer.

The only requirement to use that with Okra is to take note of "ALB Listener ARN", which is used to tell Okra 
which loadbalancer to use for traffic management.

### Provision Kubernetes Clusters

Once `okra` is ready and you see no error, add one or more EKS clusters on your AWS account.

- Tag your EKS clusters with `Service=demo`, as we use it to let `okra` auto-import those as ArgoCD cluster secrets.
- Create one or more target groups per EKS cluster and take note of target group ARNS

### Deploy Applications onto Clusters

Do either of the below to register clusters to ArgoCD and Okra

- Run `argocd cluster add` on the new cluster and either (1) create a new ArgoCD `Application` custom resource per cluster or (2) let ArgoCD `ApplicationSet` custom resource to auto-deploy onto the clusters
- Use Okra's `ClusterSet` to auto-import EKS clusters to ArgoCD and use `ApplicationSet` to auto-deploy

See [Argocd cluster add - Argo CD - Declarative GitOps CD for Kubernetes](https://argo-cd.readthedocs.io/en/stable/user-guide/commands/argocd_cluster_add/) for more information on the `argocd cluster add` command.

Also see [argoproj-labs/applicationset](https://github.com/argoproj-labs/applicationset) for more information on ArgoCD `ApplicationSet` and the controller.

#### Auto-Deploy with ApplicationSet and ClusterSet

Assuming your Okra instance has access to AWS EKS and STS APIs, you can use Okra's `ClusterSet` custom resources to
auto-discover EKS clusters and create corresponding ArgoCD cluster secrets.

This, in combination with ArgoCD `ApplicationSet`, enables you to auto-deploy your applications onto any newly created
EKS clusters, without ever touching ArgoCD or Okra at all.

The following `ClusterSet` auto-discovers AWS EKS clusters tagged with `Service=demo` and creates
corresponding ArgoCD cluster secrets.

```yaml
apiVersion: okra.mumo.co/v1alpha1
kind: ClusterSet
metadata:
  name: cell1
spec:
  generators:
  - awseks:
      selector:
        matchTags:
          Service: demo
  template:
    metadata:
      labels:
        service: demo
```

Note that `template.metadat.labels.sevice` instruct cluster secrets to get `metadata.labels` of `service: demo`, so that `AWSTargetGroupSet` can discover those clusters by labels.

Let's say you had an EKS cluster that looks like the below:

```
$ aws eks describe-cluster --name cdk1
{
    "cluster": {
        "name": "cdk1",
        "arn": "arn:aws:eks:REGION:ACCOUNT:cluster/cdk1",
        "createdAt": "2021-09-20T03:21:44.391000+00:00",
        "version": "1.21",
        "endpoint": "https://SOME_CLUSTER_ID.SOME_SHARD_ID.REGION.eks.amazonaws.com",
        "roleArn": "arn:aws:iam::ACCOUNT:role/NAME",
        "resourcesVpcConfig": {
            "subnetIds": [
                "subnet-aaa",
                "subnet-bbb",
                "subnet-ccc"
            ],
            "securityGroupIds": [
                "sg-ddd"
            ],
            "clusterSecurityGroupId": "sg-eee",
            "vpcId": "vpc-fff",
            "endpointPublicAccess": true,
            "endpointPrivateAccess": true,
            "publicAccessCidrs": [
                "0.0.0.0/0"
            ]
        },
        "kubernetesNetworkConfig": {
            "serviceIpv4Cidr": "172.20.0.0/16"
        },
        "logging": {
            "clusterLogging": [
                {
                    "types": [
                        "api",
                        "audit",
                        "authenticator",
                        "controllerManager",
                        "scheduler"
                    ],
                    "enabled": false
                }
            ]
        },
        "identity": {
            "oidc": {
              ...
            }
        },
        "status": "ACTIVE",
        "certificateAuthority": {
          ...
        },
        "platformVersion": "eks.2",
        "tags": {
            "Service": "demo"
        }
    }
}
```

Note that this Okra is able to find this EKS cluster because:

- This cluster has `tags` of `"Service": "demo"` while
- The ClusterSet created above has `generators[].awseks.selector.matchTags` of `Service: demo`

An ArgoCD cluster secret created by the above `ClusterSet` should look the below, which is a regular ArgoCD cluster secret with the specified labels.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cdk1
  namespace: default
  labels:
    argocd.argoproj.io/secret-type: cluster
    service: demo
type: Opaque
data:
  config: <BASE64 ENCODED CONFIG JSON>
  name: <BASE64 ENCODED CLUSTER NAME>
  server: <BASE64 ENCODED HTTPS URL OF K8S API ENDPOINT>
```

Again, note that this clustser secret got `metadata.labels` of `service: demo`, because `ClusterSet` had `template.metadat.labels.sevice`.

### Register Target Groups

Okra works by gradually updating target groups weights behind a loadbalancer. In order to do so,
you firstly need to tell which target groups to manage, by creating `AWSTargetGroup` custom resource
on your management cluster per target group.

An `AWSTargetGroup` custom resource is basically a target group ARN with a version number and labels.

```yaml
apiVersion: okra.mumo.co/v1alpha1
  kind: AWSTargetGroup
  metadata:
    name: default-web1
    labels:
      role: web
      okra.mumo.co/version: 1.0.0
  spec:
    # Replace REGION, ACCOUNT, NAME, and ID with the actual values
    arn: arn:aws:elasticloadbalancing:REGION:ACCOUNT:targetgroup/NAME/ID
```

#### Auto-Register Target Groups with AWSTargetGroupSet

Assuming you've already created ArgoCD cluster secret for clusters, Okra's `AWSTargetGroupSet` can be used to auto-discover
target groups associated to the cluster and register those as `AWSTargetGroup` resources.

The following `AWSTargetGroupSet` auto-discovers `TargetGroupBinding` resources labeled with `role=web` from clusters
labeled with `service=demo`, to create corresponding `AWSTargetGroup` resources in the management cluster.

```yaml
apiVersion: okra.mumo.co/v1alpha1
kind: AWSTargetGroupSet
metadata:
  name: cell1
  namespace: default
spec:
  generators:
  - awseks:
      bindingSelector:
        matchLabels:
          role: web
      clusterSelector:
        matchLabels:
          service: demo
  template:
    metadata: {}
```

Let's say you had the below `TargetGroupBinding` custom resource labled with `role: web` in the new cluster labeled with `serivce: demo`:

```yaml
# In the new cluster
apiVersion: elbv2.k8s.aws/v1beta1
kind: TargetGroupBinding
metadata:
  name: web1
  namespace: default
  labels:
    role: web
    okra.mumo.co/version: 1.0.0
spec:
  arn: arn:aws:elasticloadbalancing:REGION:ACCOUNT:targetgroup/NAME/ID
```

Okra is able to find the cluster thanks to `clusterSelector.matchLabels.service=demo` and also able find this target group binding thanks to `bindingSelector.matchLabels.role=web`.

The outcome is that Okra creates the below `AWSTargetGroup` in the management cluster. Note that `metadata.name` of it is
derived from the original `TargetGroupBinding`'s `metadata.namespace` and `metadata.name`, concatenated with `-` in between.

```yaml
# In the management cluster
apiVersion: okra.mumo.co/v1alpha1
  kind: AWSTargetGroup
  metadata:
    name: default-web1
    labels:
      role: web
      okra.mumo.co/version: 1.0.0
  spec:
    # Replace REGION, ACCOUNT, NAME, and ID with the actual values
    arn: arn:aws:elasticloadbalancing:REGION:ACCOUNT:targetgroup/NAME/ID
```

The label `role: web` is later used by `Cell` to detect it as a candidate for a canary target group, and the `okra.mumo.co/version: 1.0.0` label is used to group and sort all the detected target groups to finally see which set of target groups are considered as a part of the next canary.

### Create Cell

Finally, create a `Cell` resource.

It specifies how it utilizes an existing AWS ALB in `Spec.Ingress.AWSApplicationLoadBalancer` and which listener rule to be used for rollout, and the information to detect target groups that serves your application.

An example `Cell` custom resource follows.

On each reconcilation loop, Okra looks for `AWSTargetGroup` resources labeled with `role=web`,
and group those up by the version numbers saved under the `okra.mumo.co/version` labels.

As `Spec.Replicas` being set to `2`, it waits until 2 latest target groups appear, and starts a canary rollout only after that.

If your application is not that big and a single cluster suffices, you can safely set `replicas: 1` or omit `replicas` at all.

```yaml
kind: Cell
metadata:
  name: cell1
spec:
  ingress:
    type: AWSApplicationLoadBalancer
    awsApplicationLoadBalancer:
      listener:
        rule:
          forward: {}
          hosts:
          - example.com
          priority: 10
      listenerARN: arn:aws:elasticloadbalancing:ap-northeast-1:ACCOUNT:listener/app/...
      targetGroupSelector:
        matchLabels:
          role: web
  replicas: 2
  updateStrategy:
    canary:
      steps:
      - setWeight: 20
      - analysis:
          args:
          - name: service-name
            value: exampleapp
          templates:
          - templateName: success-rate
      - pause:
          duration: 5s
      - setWeight: 40
    type: Canary
```

#### Configuring ALB listener rule for the Cell

The listener rule part is required in order to configure your ALB.

```yaml
listener:
  rule:
    forward: {}
    hosts:
    - example.com
    priority: 10
```

And this directly corresponds to the configuration of an [ALB Listener Rule](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/listener-update-rules.html).

`priority: 10` is the priority of the listener rule to be added to your ALB listener and `hosts: [example.com]` is the conditions associated to the rule.

ALB supports both default and non-default listener rules. Every non-default listener rule requires a priority and non-empty rule conditions.

Okra is designed to not modify the default rule as it can be disruptive sometimes. That's why it requires `priority` and a rule condition.

Other rule conditions like `headers`, `methods`, `pathPatterns`, and so on are also supported. See the output of `kubectl explain awsapplicationloadbalancerconfig.spec.listener --recursive` for all the available conditions.

#### Configuring Canary Rollout Steps for the Cell

`spec.updateStategy.canary.steps` contains a definition of canary rollout steps.

Each step can be any of the belows:

- `setWeight`: updates the canary target groups total weight to the given value. For example, when there's only one canary target group to be rolled out and it's `setWeight: 20`, the canary target group gets weight of `20`. If there were two canary target groups, each gets weight of `10`
- `analysis`: runs ArgoCD `AnalysisRun` with given arguments. See [ArgoCD's documentation on Analysis](https://argoproj.github.io/argo-rollouts/features/analysis/#background-analysis) for more information on `AnalysisRun` and its template `AnalysisTemplate`.
- `pause`: pauses the rollout for the duration.

### Create and Rollout New Clusters

Now you're all set!

Every time you provision new clusters with greater version number, `Cell` automatically discovers new target groups associated to the new clusters, gradually update loadbalancer target groups weights while running various analysis.

Need a Kubernetes version upgrade? Create new Kubernetes clusters with the new Kubernetes version and watch `Cell` automatically and safely rolls out the clusters.

Need a host OS upgrade? Create new clusters with nodes with the new version of the host OS and watch `Cell` rolls out the new clusters.

And you can do the same on every kind of cluster-wide change! Enjoy running your ephemeral Kubernetes clusters.

### Analysises and Experiments

`Okra` provides almost the same features for Argo Rollouts [Analysises](https://argoproj.github.io/argo-rollouts/features/analysis/) and [Experiments](https://argoproj.github.io/argo-rollouts/features/experiment/).

A major difference between Argo Rollouts' and Okra's is that Okra's provides access to the `status.desiredVersion` field in an analysis and experiment query argument, so that you can analyze and experiment your canary based on metrics specific to the version of your clusters.

In the below example, we have an analysis step that creates an analaysis run from a analysis template named `success-rate`, whose argument `cluster-version` is set to the value obtained from `cell.status.desiredVersion`.

The `desiredVersion` status field contains the desired version number(obtained from e.g. EKS cluster tags and AWS target group tags) of the cluster repliacs being rolled out, so that you can analyze based on metrics specific to the newly rolled out clusters.

A typical cell whose one of canary steps is a `analysis` would look like the below. Notice the `fieldPath: status.desiredVersion` used to dynamically generate the `cluster-version` analysis run argument.

```yaml
apiVersion: okra.mumo.co/v1alpha1
kind: Cell
metadata:
  name: web
spec:
  updateStrategy:
    type: Canary
    canary:
      steps:
      # ...
      - analysis:
          templates:
          - templateName: success-rate
          args:
          - name: service-name
            value: guestbook-svc.default.svc.cluster.local
          - name: cluster-version
            valueFrom:
              fieldRef:
                fieldPath: status.desiredVersion
```

Similarly, an experiment step can inculde a `fieldPath` to have a dynaically generate argument:

```yaml
apiVersion: okra.mumo.co/v1alpha1
kind: Cell
metadata:
  name: web
spec:
  updateStrategy:
    type: Canary
    canary:
      steps:
      # ...
      - experiment:
          duration: 5m
          templates:
          - name: wy
            # references the wy replicaset defined below
            specRef: wy
            # This should default to 1 as defined by Argo Rollouts but
            # the author observed that it doesn't work in practice.
            # 
            replicas: 1
          analyses:
          - name: success-rate-dd
            templateName: success-rate-dd
            args:
            - name: service-name
              value: wy-serve
            - name: cluster-version
              valueFrom:
                fieldRef:
                  fieldPath: status.desiredVersion
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  labels:
    app: wy
  name: wy
spec:
  replicas: 0
  selector:
    matchLabels:
      app: wy
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: wy
    spec:
      containers:
      - image: mumoshu/wy:latest
        name: wy
        ports:
        - containerPort: 8080
        resources: {}
        args:
        - repeat
        - get
        - -forever
        - -interval=5s
        - -url=http://localhost:8080
        - -argocd-cluster-secret=cdk1
        - -service=wy-serve
        - -remote-port=8080
        - -local-port=8080
        envFrom:
        - secretRef:
            name: wy
            optional: true
```

#### Datadog

As explained earlier, `Okra` relies on Argo Rollouts `Datadog` support.
That is, you define a Okra `Cell`, so that the okra controller creates either Argo Rollouts `AnalysisRun` or `Experiment`, which in turn queries Datadog metrics with [the "Query timeseries points" API](https://docs.datadoghq.com/api/latest/metrics/#query-timeseries-points).

If you're curious how you'd instrument your app so that it's metrices cna be used from Okra, you'd better get started by reading e.g. [Mapping Prometheus Metrics to Datadog Metrics](https://docs.datadoghq.com/integrations/guide/prometheus-metrics/).

Before authoring a complex `Cell` spec including Analysis and Expriment, the author recommends you to try browsing Datadog dashboard, or use simpler tool like `curl` to query metries.
After you've done so, start tinkering with Okra, so that when it break you can be extra sure when and where it broke!

## Notes

It is inteded to be deployed onto a "control-plane" cluster to where you usually deploy applications like ArgoCD.

It requires you to use:

- NLB or ALB to load-balance traffic "across" clusters
  - You bring your own LB, Listener, and tell `okra` the Listener ID, Number of Target Groups per Cell, and a label to group target groups by version.
- Uses ArgoCD ApplicationSets to deploy your applications onto cluster(s)

In the future, it may add support for using Route 53 Weighted Routing instead of ALB.

Although we assume you use ApplicationSet for app deployments, it isn't really a strict requirement. Okra doesn't communiate with ArgoCD or ApplicationSet. All Okra does is to discover EKS clusters, create and label target groups for the discovered clusters, and rollout the target groups. You can just bring your own tool to deploy apps onto the clusters today.

It supports complex configurations like below:

- One or more clusters per cell, or an ALB listener rule. Imagine a case that you need a pair of clusters to serve your service. `okra` is able to canary-deploy the pair of clusters, by periodically updating two target group weights as a whole.

The following situations are handled by Okra:

- When there are enough number of "new" target groups, Okra gradually updates target group weights for a rollout
- Okra automatically falls back to a "old" target groups when there are only old target groups in the AWS account while ALB points to "new" target groups that disappeared

## CRDs

`Okra` provides several Kuberntetes CustomResourceDefinitions(CRD) to achieve its goal.

See [crd.md](/docs/crd.md) for more documentation and details of each CRD.

## CLI

`okra` provides 3 executables.

- `okrad`: the Kubernetes controller manager that consists of various Kubernetes controller for Okra CRDs. Intended to be run in a Kubernetes cluster.
- `okractl`: A `kubectl`-like CLI application that is for interacting with `okrad` through Kubernetes API server. Intended to be run on your machine or on a CI system for automation.
- `okra`: the standalone CLI application that does its best to provide every single logic implemented in `okrad`'s controllers. Intended to be run in CI to replicate `okrad`'s functionality on a CI system, or to test each okra functionality in isolation.

The standard and author's recommended usage of Okra involves `okrad` and `okractl`.

For `okra`, we do our best to expose every single `okrad` + `okractl` functionality via respective `okra` CLI commands, so that you can test each functionality in isolation.

It may be even possible to build your own CI job that replaces `okra` out of those commands!

See [CLI](/docs/cli.md) for more information and its usage.

## Related Projects

Okra is inspired by various open-source projects listed below.

- [ArgoCD](https://argoproj.github.io/argo-cd/) is a continuous deployment system that embraces GitOps to sync desired state stored in Git with the Kubernetes cluster's state. `okra` integrates with `ArgoCD` and especially its `ApplicationSet` controller for applicaation deployments.
  - `okra` relies on ArgoCD `ApplicationSet` controller's [`Cluster Generator` feature](https://argocd-applicationset.readthedocs.io/en/stable/Generators/#label-selector)
- [Flagger](https://flagger.app/) and [Argo Rollouts](https://argoproj.github.io/argo-rollouts/) enables canary deployments of apps running across pods. `okra` enables canary deployments of clusters running on IaaS.
- [argocd-clusterset](https://github.com/mumoshu/argocd-clusterset) auto-discovers EKS clusters and turns those into ArgoCD cluster secrets. `okra` does the same with its [`ClusterSet` CRD](/crd.md#clusterset) and `argocdcluster-controller`.
- [terraform-provider-eksctl's courier_alb resource](https://github.com/mumoshu/terraform-provider-eksctl/tree/master/pkg/courier) enables canary deployments on target groups behind AWS ALB with metrics analysis for Datadog and CloudWatc metrics. `okra` does the same with it's [`AWSApplicationLoadBalancerConfig` CRD](/crd.md#awsapplicationloadbalancerconfig) and `awsapplicationloadbalancerconfig-controller`.

## Why is it named "okra"?

Initially it was named `kubearray`, but the original author wanted something more catchy and pretty.

In the beginning of this project, the author thought that hot-swapping a cluster while keeping your apps running looks like hot-swaping a drive while keeping a server running.

We tend to call a cluster of storages where each storage drive can be hot-swapped a "storage array", hence calling a tool to build a cluster of clusters where each cluster can be hot-swapped "kubearray" seemed like a good idea.

Later, he searched over the Internet for a prettier and catchier alternative. While browsing a list of cool Japanese terms with 3 syllables, he encountered "okra". "Okra" is a pod vegetable full of edible seeds. The term is relatively unique that it sounds almost the same in both Japanese and English. The author thought that "okra" can be a good metaphor for a cluster of sub-clusters when each seed in an okra is compared to a sub-cluster.
