<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-start-->
!!! tip ""
    === "deploy"
          KubeStellar's helm chart automatically creates a Workload Management
          Workspace (WMW) for you to store kubernetes workload descriptions and KubeStellar
          control objects in. The automatically created WMW is at `root:wmw1`.

          Create an EdgePlacement control object to direct where your workload runs using the 'location-group=edge' label selector. This label selector's value ensures your workload is directed to both clusters, as they were labeled with 'location-group=edge' when you issued the 'kubestellar prep-for-cluster' command above.

          In the `root:wmw1` workspace create the following `EdgePlacement` object: 
            
          ```shell linenums="1" hl_lines="10 11 16 21"
          export KUBECONFIG=ks-core.kubeconfig
          kubectl ws root:wmw1

          kubectl apply -f - <<EOF
          apiVersion: edge.kubestellar.io/v2alpha1
          kind: EdgePlacement
          metadata:
            name: my-first-edge-placement
          spec:
            locationSelectors:
            - matchLabels: {"location-group":"edge"}
            downsync:
            - apiGroup: ""
              resources: [ configmaps ]
              namespaceSelectors:
              - matchLabels: {"common":"sure-is"}
              objectNames: [ "*" ]
            - apiGroup: apps
              resources: [ deployments ]
              namespaceSelectors:
              - matchLabels: {"common":"sure-is"}
              objectNames: [ my-first-kubestellar-deployment ]
            - apiGroup: apis.kcp.io
              resources: [ apibindings ]
              namespaceSelectors: []
              objectNames: [ "bind-kubernetes", "bind-apps" ]
          EOF
          ```

          check if your edgeplacement was applied to the **ks-core** `kubestellar` namespace correctly
          ```shell
          export KUBECONFIG=ks-core.kubeconfig
          kubectl ws root:wmw1
          kubectl get edgeplacements -n kubestellar -o yaml
          ```

          Now, apply the HTTP server workload definition into the WMW on **ks-core**. Note the namespace label matches the label in the namespaceSelector for the EdgePlacement (`my-first-edge-placement`) object created above. 


          ```shell linenums="1" hl_lines="7 14 29"
          export KUBECONFIG=ks-core.kubeconfig
          kubectl apply -f - <<EOF
          apiVersion: v1
          kind: Namespace
          metadata:
            name: my-namespace
            labels: {common: "sure-is"}
          ---
          apiVersion: v1
          kind: ConfigMap
          metadata:
            namespace: my-namespace
            name: httpd-htdocs
            labels: {common: "sure-is"}
          data:
            index.html: |
              <!DOCTYPE html>
              <html>
                <body>
                  This web site is hosted on edge-cluster1 and edge-cluster2.
                </body>
              </html>
          ---
          apiVersion: apps/v1
          kind: Deployment
          metadata:
            namespace: my-namespace
            name: my-first-kubestellar-deployment
            labels: {common: "sure-is"}
          spec:
            selector: {matchLabels: {app: common} }
            template:
              metadata:
                labels: {app: common}
              spec:
                containers:
                - name: httpd
                  image: library/httpd:2.4
                  ports:
                  - name: http
                    containerPort: 80
                    hostPort: 8081
                    protocol: TCP
                  volumeMounts:
                  - name: htdocs
                    readOnly: true
                    mountPath: /usr/local/apache2/htdocs
                volumes:
                - name: htdocs
                  configMap:
                    name: httpd-htdocs
                    optional: false
          EOF
          ```

          check if your configmap and deployment was applied to the **ks-core** `my-namespace` namespace correctly
          ```shell
          export KUBECONFIG=ks-core.kubeconfig
          kubectl ws root:wmw1
          kubectl get deployments/my-first-kubestellar-deployment -n my-namespace -o yaml
          kubectl get deployments,cm -n my-namespace
          ```

          Now, let's check that the deployment was created in the kind **edge-cluster1** cluster (it may take up to 30 seconds to appear):

          ```shell
          KUBECONFIG=~/.kube/config kubectl --context kind-edge-cluster1 get deployments -A
          ```

          you should see output including:

          ``` { .sh .no-copy }
          NAMESPACE           NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
          my-namespace        my-first-kubestellar-deployment    1/1        1            1       6m48s
          ```

          And, check the `edge-cluster2` kind cluster for the same:

          ```shell
          KUBECONFIG=~/.kube/config kubectl --context kind-edge-cluster2 get deployments -A
          ```

          you should see output including:

          ``` { .sh .no-copy }
          NAMESPACE           NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
          my-namespace        my-first-kubestellar-deployment    1/1        1            1       7m54s
          ```

          Finally, let's check that the workload is working in both clusters:

          For **edge-cluster1**:

          ```shell
          export KUBECONFIG=~/.kube/config
          while [[ $(kubectl --context kind-edge-cluster1 get pod \
            -l "app=common" -n my-namespace \
            -o jsonpath='{.items[0].status.phase}') != "Running" ]]; \
            do sleep 5; done;curl http://localhost:8094
          ```

          you should see the output:

          ```html
          <!DOCTYPE html>
          <html>
            <body>
              This web site is hosted on edge-cluster1 and edge-cluster2.
            </body>
          </html>
          ```

          For **edge-cluster2**:

          ```shell
          export KUBECONFIG=~/.kube/config
          while [[ $(kubectl --context kind-edge-cluster2 get pod \
            -l "app=common" -n my-namespace \
            -o jsonpath='{.items[0].status.phase}') != "Running" ]]; \
            do sleep 5; done;curl http://localhost:8096
          ```

          you should see the output:

          ```html
          <!DOCTYPE html>
          <html>
            <body>
              This web site is hosted on edge-cluster1 and edge-cluster2.
            </body>
          </html>
          ```
    === "uh oh, error?"
          If you are unable to see the namespace 'my-namespace' or the deployment 'my-first-kubestellar-deployment' you can view the logs for the KubeStellar syncer on the **edge-cluster1** kind cluster:

          ```
          export KUBECONFIG=~/.kube/config 
          kubectl config use-context kind-edge-cluster1
          ks_ns_edge_cluster1=$(kubectl get namespaces -o custom-columns=:metadata.name | grep 'kubestellar-')

          kubectl logs pod/$(kubectl get pods -n $(ks_ns_edge_cluster1) -o custom-columns=:metadata.name | grep 'kubestellar-') -n $(ks_ns_edge_cluster1)
          ```

          and on the **edge-cluster2** kind cluster:

          ```
          export KUBECONFIG=~/.kube/config 
          kubectl config use-context kind-edge-cluster2
          ks_ns_edge_cluster2=$(kubectl get namespaces -o custom-columns=:metadata.name | grep 'kubestellar-')
          kubectl logs pod/$(kubectl get pods -n $(ks_ns_edge_cluster2) -o custom-columns=:metadata.name | grep 'kubestellar-') -n $(ks_ns_edge_cluster2)
          ```

          If you see a `connection refused` error in either KubeStellar Syncer log(s):

          `E1021 21:22:58.000110       1 reflector.go:138] k8s.io/client-go@v0.0.0-20230210192259-aaa28aa88b2d/tools/cache/reflector.go:215: Failed to watch *v2alpha1.EdgeSyncConfig: failed to list *v2alpha1.EdgeSyncConfig: Get "https://kubestellar.core:1119/apis/edge.kubestellar.io/v2alpha1/edgesyncconfigs?limit=500&resourceVersion=0": dial tcp 127.0.0.1:1119: connect: connection refused`

          it means that your `/etc/hosts` does not have a proper ip address (NOT `127.0.0.1`) listed for the `kubestellar.core` hostname. Once there is a valid address in `/etc/hosts` for `kubestellar.core`, the syncer will begin to work properly and pull the namespace, deployment, and configmap from this instruction set. 

<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-end-->
