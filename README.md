# Okra

`Okra` is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/) and a set of [CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) which provide advanced multi-cluster appilcation rollout capabilities, such as canary deployment of clusters.

`okra` manages **Cell** for you. A cell is like a storage array but for Kubernetes clusters.

You hot-swap a disk in a storage array while running. Similarly, with `okra` you hot-swap a cluster in a system while running.

`okra` (currently) integrates with AWS ALB and NLB and target groups for traffic management, CloudWatch Metrics and Datadog for canary analysis.

## Goals

`okra` eases managing ephemeral Kubernetes clusters.

If you've been using ephemeral Kubernetes clusters and employed blue-green or canary deployments for zero-downtime cluster updates, you might have suffered from a lot of manual steps required. `okra` is intended to automate all those steps.

In the best scenario, a system update with `okra` looks like the below.

- You provision one or more new clusters with some cluster tags
- An external system like ArgoCD with ApplicationSet deploys your apps to the new clusters
- `cell-controller` starts discovering the new clusters by tags
- Once there are enough clusters, `cell-controller` starts updating the loadbalancer configuration to gradually migrate traffic from the old to the new clusters.
- Have some coffee. `okra` will run various steps to ensure there are no errors, and it reverts the loadbalancer configuration changes when there are too many errors or test failures.

## Project Status and Scope

`okra` currently works on AWS only, but the design and implementation is generic enough to be capable of adding more IaaS supports. Any contribution around that is welcomed.

## How does it work?

Okra's `cell-contorller` will manage the traffic shift across clusters.

In the beginning, a user gives each `Cell` a set of settings to discover AWS target groups and configure loadbalancers, and metrics.

The controller periodically discovers AWS target groups. Once there are enough number of new target groups, it then compares the target groups associated to the loadbalancer. If there's any difference, it starts updating the ALB while checking various metrics for safe rollout.

Okra uses Kubernetes CRDs and custom resources as a state store and uses the standard Kubernetes API to interact with resources.

Okra calls various AWS APIs to create and update AWS target groups and update AWS ALB and NLB forward config for traffic management.

### Comparison with Flagger and Argo Rollouts

Unlike `Argo Rollouts` and `Flagger`, in `Okra` there is no notions of "active" and "preview" services for a blue-green deployment, or "canary" and "stable" services for a canary deployment.

It assumes there's one or more target groups per cell. `cell` basically does a canary deployment, where the old set of target groups is consdidered "stable" and the new set of target groups is considered "canary".

In `Flagger` or `Argo Rollouts`, you need to update its K8s resource to trigger a new rollout. In Okra you don't need to do so. You preconfigure its resource and Okra auto-starts a rollout once it discovers enough number of new target groups.

## Concepts

`okra` updates your `Cell`.

A okra `Cell` is composed of target groups and an AWS loadbalancer, and a set of metrics for canary anlysis.

Each target group is tied to a `cluster`, where a `cluster` is a Kubernetes cluster that runs your container workloads.

An `application` is deployed onto `clusters` by `ArgoCD`. The traffic to the `application` is routed via an [AWS ALB](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/introduction.html) in front of `clusters`.

```
kind: Cell
spec:
  ingress:
    type: AWSApplicationLoadBalancer
    awsApplicationLoadBalancer:
      listenerARN: ...
      targetGroupSelector:
        matchLabels:
          role: web
  replicas: 2
  versionedBy:
    label: "okra.mumoshu.github.io/version"
```

`okra` acts as an application traffic migrator.

It detects new `target groups`, and live migrate traffic by hot-swaping old target groups serving the affected `applications` with the new target groups, while keepining the `applications` up and running.

## Usage

It is inteded to be deployed onto a "control-plane" cluster to where you usually deploy applications like ArgoCD.

It requires you to use:

- NLB or ALB to load-balance traffic "across" clusters
  - You bring your own LB, Listener, and tell `okra` the Listener ID, Number of Target Groups per Cell, and a label to group target groups by version.
- Uses ArgoCD ApplicationSets to deploy your applications onto cluster(s)

In the future, it may add support for using Route 53 Weighted Routing instead of ALB.

Although we assume you use ApplicationSet for app deployments, it isn't really a strict requirement. Okra doesn't communiate with ArgoCD or ApplicationSet. All Okra does is to discover EKS clusters, create and label target groups for the discovered clusters, and rollout the target groups. You can just bring your own tool to deploy apps onto the clusters today.

It supports complex configurations like below:

- One or more clusters per cell, or an ALB listener rule. Imagine a case that you need a pair of clusters to serve your service. `okra` is able to canary-deploy the pair of clusters, by periodically updating two target group weights as a whole.

A complex example of Cell would look like the below:

```
apiVersion: okra.mumoshu.github.io/v1alpha1
kind: Cell
metadata:
  name: mysystem
spec:
  clusters:
  - name: mycluster-web-1-v2
    eks:
      #clusterName: mycluster-web-1-v2
    # by default this fetches the EKS cluster named mycluster-v2
    # This waits until the cluster is ready and running
    requireReady: true
    # or allow notready for 60m...
    #readyTimeout: 60m
    labels:
      role: web
  - name: mycluster-web-2-v2
    eks: {}
    labels:
      role: web
  - name: mycluster-api-v1
    eks: {}
    labels:
      role: api
      targetGroupARN: ...someARN...
  - name: mycluster-backgroundjobs-v1
    eks: {}
    labels:
      role: backgroundjobs
  applications:
  - name: web
    # selector is required when there are two or more clusters
    clusterSelector:
      matchEKSTags:
        role: web
    # This automatically waits for related applicationset(s) to be reconciled
    argocd:
      clusterSecret:
        labels:
          myservice-role: web
          # secret-type label is automatically added.
          #argocd.argoproj.io/secret-type: cluster
      applicationSet:
        name: ...
        #selector: ...
      updateStrategy:
        type: CreateBeforeDelete
    # There's implicit ordering between argocdClusterSecret and alb updates.
    # okra will firstly update argocdClusterSecret to add new clusters,
    # and then updates ALB so that it routes more traffic to new clusters.
    # Once all the checks passed, it finally removes old clusters from
    # argocdClusterSecret.
    # This is how argocdClusterSecret.updateStrategy=CreateBeforeDelete works in concert with alb
    alb:
      listenerARN: ...
      forwardConfig:
        priority: 10
        targetGroup:
          # this doesn't work when selector matched two or more clusters
          arn: ...
          #targetgroupBindingSelector: svc=web
          #arnFromLabel: targetGroupARN
      updateStrategy: #canary
        stepWeight: 10
        totalWeight: 100
    checks:
    - name: dd
      analysis:
        query: |
          ... {{.Vars.clusterName}}
          ... {{.Vars.albListenerARN}} ...{{.Vars.targetGroupARN}}
        max: 0.1
  - name: api
    clusterSelector:
      role: api
    argocd:
      clusterSecret:
        labels:
          myservice-role: api
          # secret-type label is automatically added.
          #argocd.argoproj.io/secret-type: cluster
      applicationSet:
        selector: ...
      updateStrategy:
        type: CreateBeforeDelete
    alb:
      listenerARN: ...
      forwardConfig:
        priority: 11
        targetGroup:
          #arn: ...
          #targetgroupBindingSelector: svc=api
          arnFromLabel: targetGroupARN
      updateStrategy: #blue-green
        stepWeight: 100
        totalWeight: 100
    checks:
    - name: dd
      analysis
        query: |
          ... {{.Vars.eksClusterName}}
          ... {{.Vars.albListenerARN}} ...{{.Vars.targetGroupARN}}
        max: 0.1
  - name: backgroundjobs
    clusterSelector:
      role: backgroundjobs
    # This automatically waits for related applicationset(s) to be reconciled
    # This result in removing old cluster secrets.
    argocd:
      clusterSecret:
        lables:
          myservice-role: backgroundjobs
      applicationSet:
        selector: ...
      updateStrategy:
        type: DeleteBeforeCreate
    # Note that argocdClusterSecret.updateStrategy=DeleteBeforeCreate with alb is not allowed(validation error).
    
    # Wait for prechecks to pass before updating the secret
    prechecks:
    - name: run-test
      job:
        template:
          spec:
            image: ...
            command: ...
            args: ...
status:
  phase: Failed
  #phase: Reverting
  #phase: FailedReverting
  applications:
  - name: web
    phase: Synced
  - name: api
    phase: Syncing
  - name: backgroundjobs
    phase: Pending
  - name: someotherservice
    phase: PrecheckFailed
    message: run-test failed. Waiting 60 seconds before starting a rollback
```

## How it works

`okra` waits for all the clusters are ready, and starts migrating workloads only after that. This ensures that there will be no traffic flapping. In the example above, `okra` starts migration only after all the clusters:

- mycluster-web-1-v2
- mycluster-web-2-v2
- mycluster-api-v1
- mycluster-backgroundjobs-v1

are up and running.

If it were to auto-discover e.g. `mycluster-web-1-v2` and start migrating workloads to it, what should it do when it discovers `mycluster-web-2-v2` next? It might work differently depending on in which timing it discovered a cluster in a same group.

Forcing to wait for all the relevant clusters to become ready before migrating workloads makes the whole story simple and deterministic. This is also a good practice to minimize downtime migrating aapps like Kafka consumers across clusters, because it may result in less resharding.

On each reconcilation loop, `okra` runs the following steps:

- Iterate over `clusters` and ensure all the clusters are ready
- Iterate over `applications` and ensure all the services have desired clusters
  - If not, for each out-of-sync service:
    - Phase=Init: Run precheck if any, recording check results paired with current ALB forward config hash. Proceed only if it succeeded. Requeue if failed. Requeue if processed. If the hash doesn't change and the recorded check results indicate successful checks, do nothing.
    - Phase=Migrating: Run N (=stepWeight / totalWeight) steps to gradually change ALB settings
      - On each step, run all the checks with retries. Record the check results in ClusterSet statusProceed only if it suceeded. Requeue if processed, so that one reconcilation loop runs only one series of checks, to avoid long blocking.
    - Phase=Completing: Run postchecks and record the results. Requeue if processed.
    = Phase=Error: If checks or prechecks failed, it will either stop or rollback. For rollback, it reads old ClusterSet revisions from the cluster.
    = Phase=Completed: Does nothing.

## Implementation

`okra` provides the following Kubernetes CRDs:

- `Check`

### `Check`

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

## Related Projects

Okra is inspired by various open-source projects listed below.

- [ArgoCD](https://argoproj.github.io/argo-cd/) is a continuous deployment system that embraces GitOps to sync desired state stored in Git with the Kubernetes cluster's state. `okra` integrates with `ArgoCD` and especially its `ApplicationSet` controller for applicaation deployments.
  - `okra` relies on ArgoCD `ApplicationSet` controller's [`Cluster Generator` feature](https://argocd-applicationset.readthedocs.io/en/stable/Generators/#label-selector)
- [Flagger](https://flagger.app/) and [Argo Rollouts](https://argoproj.github.io/argo-rollouts/) enables canary deployments of apps running across pods. `okra` enables canary deployments of clusters running on IaaS.
- [argocd-clusterset](https://github.com/mumoshu/argocd-clusterset) auto-discovers EKS clusters and turns those into ArgoCD cluster secrets. `okra` does the same with its `ArgoCDCluster` CRD and `argocdcluster-controller`.
- [terraform-provider-eksctl's courier_alb resource](https://github.com/mumoshu/terraform-provider-eksctl/tree/master/pkg/courier) enables canary deployments on target groups behind AWS ALB with metrics analysis for Datadog and CloudWatc metrics. `okra` does the same with it's `ALB` CRD and `alb-controller`.

## Why is it named "okra"?

Initially it was named `kubearray`, but the original author wanted something more catchy and pretty.

In the beginning of this project, the author thought that hot-swapping a cluster while keeping your apps running looks like hot-swaping a drive while keeping a server running.

We tend to call a cluster of storages where each storage drive can be hot-swapped a "storage array", hence calling a tool to build a cluster of clusters where each cluster can be hot-swapped "kubearray" seemed like a good idea.

Then he searched over the Internet. While browsing a list of cool Japanese terms with 3 syllables, he encountered "okra". "Okra" is a pod vegetable full of edible seeds. The term is relatively unique that it sounds almost the same in both Japanese and English. The author thought that "okra" can be a good metaphor for a cluster of sub-clusters when each seed in an okra is compared to a sub-cluster.
