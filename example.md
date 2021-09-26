
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
