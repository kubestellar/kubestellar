<!--kubestellar-kind-ip-error-start-->
Did you received the following error:
```Error: Get "https://some_hostname.some_domain_name:{{config.ks_kind_port_num}}/clusters/root/apis/tenancy.kcp.io/v1alpha1/workspaces": dial tcp: lookup some_hostname.some_domain_name on x.x.x.x: no such host``

A common error occurs if you set your port number to a pre-occupied port number and/or you set your EXTERNAL_HOSTNAME to something other than "localhost" so that you can reach your <span class="Space-Bd-BT">KUBESTELLAR</span> Core from another host, check the following:
         
Check if the port specified in the **ks-core** Kind cluster configuration and the EXTERNAL_PORT helm value are occupied by another application:

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;1. is the `hostPort`` specified in the **ks-core** Kind cluster configuration is occupied by another process?  If so, delete the **ks-core** Kind cluster and create it again using an available port for your 'hostPort' value

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;2. if you change the port for your **ks-core** 'hostPort', remember to also use that port as the helm 'EXTERNAL_PORT' value

Check that your EXTERNAL_HOSTNAME helm value is reachable via DNS:

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;1. use 'nslookup <value of EXTERNAL_HOSTNAME>' to make sure there is a valid IP address associated with the hostname you have specified

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;2. make sure your EXTERNAL_HOSTNAME and associated IP address are listed in your /etc/hosts file.

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;3. make sure the IP address is associated with the system where you have deployed the **ks-core** Kind cluster

if there is nothing obvious, [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)
<!--kubestellar-kind-ip-error-end-->
