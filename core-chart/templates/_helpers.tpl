{{/*
Expand ITS PCH custeradm container.
*/}}
{{- define "pch.its.custeradm" -}}
- name: "{{"{{.HookName}}-clusteradm"}}"
  image: quay.io/kubestellar/clusteradm:{{.Values.CLUSTERADM_VERSION}}
  args:
  - init
  - -v={{.Values.verbosity.clusteradm | default .Values.verbosity.default | default 2 }}
  - --wait
  env:
  - name: KUBECONFIG
    value: "{{"/etc/kube/{{.ITSkubeconfig}}"}}"
  volumeMounts:
  - name: kubeconfig
    mountPath: "/etc/kube"
    readOnly: true
{{- end }}

{{/*
Expand ITS PCH statusaddon container.
*/}}
{{- define "pch.its.statusaddon" -}}
- name: "{{"{{.HookName}}-statusaddon"}}"
  image: quay.io/kubestellar/helm:{{.Values.HELM_VERSION}}
  args:
  - upgrade
  - --install
  - status-addon
  - oci://ghcr.io/kubestellar/ocm-status-addon-chart
  - --version
  - v{{.Values.OCM_STATUS_ADDON_VERSION}}
  - --namespace
  - open-cluster-management
  - --create-namespace
  - --set
  - "controller.verbosity={{.Values.status_controller.v | default .Values.verbosity.status_controller | default .Values.verbosity.default | default 2 }}"
  - --set
  - "agent.hub_burst={{.Values.status_agent.hub_burst}}"
  - --set
  - "agent.hub_qps={{.Values.status_agent.hub_qps}}"
  - --set
  - "agent.local_burst={{.Values.status_agent.local_burst}}"
  - --set
  - "agent.local_qps={{.Values.status_agent.local_qps}}"
  - --set
  - "agent.log_flush_frequency={{.Values.status_agent.log_flush_frequency}}"
  - --set
  - "agent.logging_format={{.Values.status_agent.logging_format}}"
  - --set
  - "agent.metrics_bind_addr={{.Values.status_agent.metrics_bind_addr}}"
  - --set
  - "agent.pprof_bind_addr={{.Values.status_agent.pprof_bind_addr}}"
  - --set
  - "agent.v={{.Values.status_agent.v | default .Values.verbosity.status_agent | default .Values.verbosity.default | default 2 }}"
  - --set
  - "agent.vmodule={{.Values.status_agent.vmodule}}"
  env:
  - name: HELM_CONFIG_HOME
    value: "/tmp"
  - name: HELM_CACHE_HOME
    value: "/tmp"
  - name: KUBECONFIG
    value: "{{"/etc/kube/{{.ITSkubeconfig}}"}}"
  volumeMounts:
  - name: kubeconfig
    mountPath: "/etc/kube"
    readOnly: true
{{- end }}
