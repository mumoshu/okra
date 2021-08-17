# hotswap

`hotswap` helps you hot-swap Kubernetes clusters while keeping your microservices up and running.

The author thought that hot-swapping a cluster while keeping your apps running looks like hot-swaping a drive while keeping a server running, hence the name `hotswap`.

## Goals

`hotswap` eases managing ephemeral Kubernetes clusters.

If you've been using ephemeral Kubernetes clusters and employed blue-green or canary deployments for zero-downtime cluster updates, you might have suffered from a lot of manual steps required. `hotswap` is intended to automate all those steps.

In the best scenario, a system update looks like the below.

- You provision one or more new clusters, and update a `System` custom resource provided by `hotswap` to include the new clusters
- Have some coffee, and `hotswap` will run various steps to safely update all the related K8s and IaaS resources. The job includes gradually migrating workloads from the old clusters to the new clusters, by updating ArgoCD configs and AWS ALB settings.

## Project Status and Scope

`hotswap` currently works on AWS only, but the design and implementation is generic enough to be capable of adding more IaaS supports. Any contribution around that is welcomed.

## Concepts

`hotswap` updates your `System`.

A `System` is composed of`clusters` and `applications`.

A `cluster` is a Kubernetes cluster that runs your container workloads.

An `application` is one of your application that is deployed onto `clusters` by `argocd`. The `application` traffic is routed via an `alb` in front of `clusters`.

```
kind: System
spec:
  clusters:
  - name: ...
  applications:
  - name: ...
    clusterSelector: ...
    argocd: ...
    alb: ...
```

`hotswap` acts as an application traffic migrator.

It detects new `clusters` in an updated `System` spec, then detects affected `applications`, live migrate traffic by hot-swaping old clusters serving the affected `applications` with the new clusters, while keepining the `applications` up and running.

## Usage

It is inteded to be deployed onto a "control-plane" cluster to where you usually deploy applications like ArgoCD.

It is opinionated in a way that it requires you to use:

- ALB(s) to load-balance traffic "across" clusters
  - You bring your own ALB, Listener, and Target Groups, and tell `hotswap` Listener ID and Target Group ARNs and Weights. If you use `aws-loadbalaner-controller`, you can use `TargetGroupBinding` only.
- Uses ArgoCD ApplicationSets to deploy your applications onto cluster(s)

In the future, it may add support for using Route 53 Weighted Routing instead of ALB and using another application deployment tool other than ArgoCD.

It supports complex configurations like below:

- One or more clusters per service=ALB listener rule. Imagine a case that you need a pair of clusters to serve your service. `hotswap` is able to canary-deploy the pair of clusters, by periodically updating two target group weights as a whole.

A complex example of System would look like the below:

```
apiVersion: hotswap.mumoshu.github.io/v1alpha1
kind: System
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
      role: web
    # This automatically waits for related applicationset(s) to be reconciled
    argocd:
      clusterSecret:
        name: web-clusters
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
    # hotswap will firstly update argocdClusterSecret to add new clusters,
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
        name: api-clusters
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
        name: backgroundjobs-clusters
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

`hotswap` waits for all the clusters are ready, and starts migrating workloads only after that. This ensures that there will be no traffic flapping. In the example above, `hotswap` starts migration only after all the clusters:

- mycluster-web-1-v2
- mycluster-web-2-v2
- mycluster-api-v1
- mycluster-backgroundjobs-v1

are up and running.

If it were to auto-discover e.g. `mycluster-web-1-v2` and start migrating workloads to it, what should it do when it discovers `mycluster-web-2-v2` next? It might work differently depending on in which timing it discovered a cluster in a same group.

Forcing to wait for all the relevant clusters to become ready before migrating workloads makes the whole story simple and deterministic. This is also a good practice to minimize downtime migrating aapps like Kafka consumers across clusters, because it may result in less resharding.

On each reconcilation loop, `hotswaps` runs the following steps:

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

`hotswap` provides the following Kubernetes CRDs:

- `System`
- `ArgoCDCluster`
- `ArgoCDApplicationDeployment`
- `Check`

### `ArgoCDCluster`

```
apiVersion: hotswap.mumoshu.github.io/v1alpha1
kind: System
metadata:
  name: mysystem
spec:
clusters:
- name: mycluster-web-1-v2
  eks:
  #  clusterName: mycluster-web-1-v2
  # by default this fetches the EKS cluster named mycluster-v2
  labels:
    role: web
# snip
```

translates to:

```
kind: ArgoCDCluster
metadata:
  name: mysystem-mycluster-web-1-v2
  labels:
    role: web
spec:
  eks:
    clusterName: mycluster-web-1-v2
status:
  ready: true
  phase: Running
```

`argocdcluster-controller` reconciles this resource, by calling AWS EKS GetCluster API, build a Kubernetes client config from it, and then writes it into a ArgoCD cluster secret named `mysystem-mycluster-web-1-v2`.

### `ArgoCDApplicationDeployment`

```
apiVersion: hotswap.mumoshu.github.io/v1alpha1
kind: System
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
# snip
  - name: web
    # selector is required when there are two or more clusters
    clusterSelector:
      role: web
    # This automatically waits for related applicationset(s) to be reconciled
    argocd:
      clusterSecret:
        name: web-clusters
        labels:
          myservice-role: web
          # secret-type label is automatically added.
          #argocd.argoproj.io/secret-type: cluster
      applicationSet:
        name: ...
        #selector: ...
      updateStrategy:
        type: CreateBeforeDelete
```

With the above, `system-controller` fetches all the cluster secrets that have corresponding items in `system.spec.clusters` by a label selector of `role=web`. The controller then puts cluster secret names into `argocdapplication.spec.clusterNames`, so that the `ArgoCDApplicationDeployment` would become:

```
kind: ArgoCDApplicationDeployment
metadata:
  name: mysystem-web
spec:
  clusterNames:
  - mysystem-mycluster-web-1-v2
  - mysystem-mycluster-web-2-v2
  clusterSecret:
    name: web-clusters
    labels:
      myservice-role: web
      # secret-type label is automatically added.
      #argocd.argoproj.io/secret-type: cluster
  applicationSet:
    selector: ...
  updateStrategy:
    type: CreateBeforeDelete
status:
  ready: true
  phase: Updating
```

With this configuration, `argocdapplicationdeployment-controller` fetches two cluster secrets `mycluster-web-1-v2` and `mycluster-web-2-v2`, concatenate the two into a cluster secret named `web-clusters`. The cluster secret is lableled `myservice-role: web` for selection by `ApplicationSet`.

ArgoCD's `ApplicationSet` controller detects the updated `web-clusters` and installs `Application`s onto the clusters according to `ApplicationSet`s.

The update strategy of `CreateBeforeDelete` results in creating updating the cluster secret to add new clusters first. It deletes the old clusters from the cluster secret only after all the `ApplicationSet` matched the `applicationSet.selector` completed deployments to the new clusters.

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
