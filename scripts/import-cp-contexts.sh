#!/usr/bin/env bash
# Copyright 2024 The KubeStellar Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -e


arg_names=""
arg_kubeconfig=""
arg_context=""
arg_host=""
arg_merge=false
arg_output=""
arg_verbose=true


display_help() {
  cat << EOF
Usage: $0 [--kubeconfig <filename>] [--context <name>] [--names <list of names>] [--replace-localhost <address>] [--merge] [-o <filename>] [-V] [-X]
--kubeconfig <filename>    use the specified kubeconfig
--context <name>           use the specified context
--names <name1>,<name2>    comma separated list of KubeFlex Control Planes names to import, instead of default *all*
--replace-localhost <host> replaces server addresses "127.0.0.1" with <host>, which can be an ip address or a domanin
--merge                    the resulting kubeconfig gets merged into the pre-existing contents of the output file
-o <filename>              specify where to write the resulting kubeconfig (- for stdout)
--silent                   no information output
-X                         enable verbose execution of the script for debugging

Note: if multiple KubeFlex Control Planes with the same names are found, only the last one found will be used in the resulting kubeconfig
EOF
}


get_kubeconfig() {
    context="$1"
    cp_name="$2"
    cp_type="$3"
    host="$4"

    echov "Getting the kubeconfig of KubeFlex Control Plane \"$cp_name\" of type \"$cp_type\" from context \"$context\":"

    # Wait for Control Plane to be ready using kubectl wait
    echov "* Waiting for KubeFlex Control Plane \"$cp_name\" in context \"$context\" to be Ready..."
    KUBECONFIG="$in_KUBECONFIG" kubectl --context "$context" wait \
        --for=condition=Ready \
        controlplane "$cp_name" \
        --timeout=300s

    # put into the shell variable "kubeconfig" the kubeconfig contents for use from outside of the hosting cluster
    if [[ "$cp_type" == "host" ]] ; then
        echov "- Using hosting cluster context \"${context}\""
        kubeconfig=$(KUBECONFIG="$in_KUBECONFIG" kubectl --context $context config view --flatten --minify)
    else
        # determine the secret name and namespace
                     key=$(KUBECONFIG="$in_KUBECONFIG" kubectl --context $context get controlplane $cp_name -o=jsonpath='{.status.secretRef.key}')
             secret_name=$(KUBECONFIG="$in_KUBECONFIG" kubectl --context $context get controlplane $cp_name -o=jsonpath='{.status.secretRef.name}')
        secret_namespace=$(KUBECONFIG="$in_KUBECONFIG" kubectl --context $context get controlplane $cp_name -o=jsonpath='{.status.secretRef.namespace}')
        # get the kubeconfig
        echov "- Using \"$key\" from \"$secret_name\" secret in \"$secret_namespace\""
        kubeconfig=$(KUBECONFIG="$in_KUBECONFIG" kubectl --context $context get secret $secret_name -n $secret_namespace -o=jsonpath="{.data.$key}" | base64 -d)
    fi

    # make kubeconfig unique for the control plane
    echov "- Making the kubeconfig unique..."
    kubeconfig=$(
        echo "$kubeconfig" \
        | yq ".clusters[0].name = \"$cp_name-cluster\"" \
        | yq ".users[0].name = \"$cp_name-admin\"" \
        | yq ".contexts[0].name = \"$cp_name\"" \
        | yq ".contexts[0].context.cluster = \"$cp_name-cluster\"" \
        | yq ".contexts[0].context.user = \"$cp_name-admin\"" \
        | yq ".current-context = \"$cp_name\""
    )

    # swap out 127.0.0.1 with an external ip address or domanin
    if [[ "$host" != "" ]] ; then
        echov "- Replacing server ip address \"127.0.0.1\" with \"$host\""
        kubeconfig=$(echo "$kubeconfig" | sed -e "s@server: https://127\.0\.0\.1:@server: https://$host:@g")
    fi

    # return the kubeconfig file contents in YAML format
    echo "$kubeconfig"
}


while (( $# > 0 )); do
    case "$1" in
    (--names|-n)
        if (( $# > 1 ));
        then { arg_names="$2"; shift; }
        else { echo "$0: missing list of KubeFlex Control Plane names" >&2; exit 1; }
        fi;;
    (--kubeconfig|-k)
        if (( $# > 1 ));
        then { arg_kubeconfig="$2"; shift; }
        else { echo "$0: missing kubeconfig filename" >&2; exit 1; }
        fi;;
    (--context|-c)
        if (( $# > 1 ));
        then { arg_context="$2"; shift; }
        else { echo "$0: missing context name" >&2; exit 1; }
        fi;;
    (--replace-localhost|-r)
        if (( $# > 1 ));
        then { arg_host="$2"; shift; }
        else { echo "$0: missing host" >&2; exit 1; }
        fi;;
    (--merge|-m)
        arg_merge=true;;
    (--output|-o)
        if (( $# > 1 ));
        then { arg_output="$2"; shift; }
        else { echo "$0: missing output filename" >&2; exit 1; }
        fi;;
    (--silent|-s)
        arg_verbose=false;;
    (-X)
        set -x;;
    (-h|--help)
        display_help
        exit 0;;
    (-*)
        echo "$0: unknown flag" >&2
        exit 1;;
    (*)
        echo "$0: unknown positional argument" >&2
        exit 1;;
    esac
    shift
done


# Define the echov function based on verbosity
if [ "$arg_verbose" == "true" ]; then
    echov() { >&2 echo "$@" ; }
else
    echov() { :; }
fi


# Determine the list of kubeconfigs to search
# based on https://kubernetes.io/docs/reference/KUBECONFIG="$in_KUBECONFIG" kubectl/generated/KUBECONFIG="$in_KUBECONFIG" kubectl_config/
echov "Determining the list of kubeconfigs to search:"
if [[ "$arg_kubeconfig" != "" ]] ; then
    in_KUBECONFIG="$arg_kubeconfig"
    if [[ "$arg_output" == "" ]] ; then
        out_kubeconfig="$arg_kubeconfig"
    fi
elif [[ "$KUBECONFIG" != "" ]] ; then
    in_KUBECONFIG="$KUBECONFIG"
    if [[ "$arg_output" == "" ]] ; then
        if [[ "$KUBECONFIG" == *":"* ]] ; then
            # hard to decide may be out_kubeconfig="$HOME/.kube/config"
            >&2 echo "ERROR: please specify an output kubeconfig filename with \"-o <filename>\"!"
            exit 1
        else
            out_kubeconfig="$KUBECONFIG"
        fi
    fi
else
    in_KUBECONFIG="$HOME/.kube/config"
    if [[ "$arg_output" == "" ]] ; then
        out_kubeconfig="$HOME/.kube/config"
    else
        out_kubeconfig="$arg_output"
    fi
fi
echov "- input kubeconfigs: \"$in_KUBECONFIG\""
echov "- output kubeconfig: \"$out_kubeconfig\""


# Determine the list of contexts to search
echov "Determining the list of contexts to search:"
if [[ "$arg_context" == "" ]] ; then
    contexts=($(KUBECONFIG="$in_KUBECONFIG" kubectl config get-contexts --no-headers -o name))
else
    contexts=("$arg_context")
fi
for context in "${contexts[@]}" ; do # for all contexts
    echov "- found context \"${context}\""
done


echov "Getting kubeconfigs of KubeFlex Control Planes:"
n=0
for context in "${contexts[@]}" ; do # for all contexts
    echov "- searching context \"${context}\""
    cps=($(KUBECONFIG="$in_KUBECONFIG" kubectl --context "${context}" get controlplanes -no-headers -o name 2> /dev/null || true))
    for i in "${!cps[@]}" ; do # for all control planes in context ${context}
        name=${cps[i]##*/}
        if [[ "$arg_names" != "" && ",$arg_names," != *",$name,"* ]] ; then
            echov "  - skipping KubeFlex Control Plane \"$name\""
            continue
        fi
        cp_name[n]=$name
        cp_type[n]=$(KUBECONFIG="$in_KUBECONFIG" kubectl get controlplane ${cp_name[n]} -o jsonpath='{.spec.type}')
        echov "  - found KubeFlex Control Plane \"${cp_name[i]}\" of type \"${cp_type[i]}\""
        cp_kubeconfig[n]=$(get_kubeconfig "${context}" "${cp_name[n]}" "${cp_type[n]}" "$arg_host")
        n=$((n+1))
    done
done


if [[ "${#cp_name[@]}" == "0" ]] ; then
    echov "No KubeFlex Control Planes found... nothing to do"
    exit 0
fi


echov "Creating new kubeconfig content with a context for each ControlPlane:"
kubeconfig_list=""
for i in "${!cp_name[@]}" ; do
    echov "- \"${cp_name[i]}\" of type \"${cp_type[i]}\" ==> saving to temporary file \"kubeconfig_${cp_name[i]}\""
    echo "${cp_kubeconfig[i]}" > "kubeconfig_${cp_name[i]}"
    kubeconfig_list="$kubeconfig_list:kubeconfig_${cp_name[i]}"
done
if [[ "$arg_merge" == "true" ]] ; then
    echov "* including current kubeconfig: \"$in_KUBECONFIG\""
    merge_KUBECONFIG="$in_KUBECONFIG$kubeconfig_list"
else
    merge_KUBECONFIG="${kubeconfig_list:1}"
fi
if [[ "$out_kubeconfig" == "-" ]] ; then
    KUBECONFIG="$merge_KUBECONFIG" kubectl config view --flatten
else
    (KUBECONFIG="$merge_KUBECONFIG" kubectl config view --flatten) > "kubeconfig_tmp"
    if [ -f "${out_kubeconfig}" ] ; then
        echov "* backing up \"${out_kubeconfig}\" to \"${out_kubeconfig}.bak\""
        mv -f "${out_kubeconfig}" "${out_kubeconfig}.bak" 2> /dev/null
    fi
    echov "* saving new kubeconfig to \"${out_kubeconfig}\""
    mv -f "kubeconfig_tmp" "${out_kubeconfig}"
fi
for i in "${!cp_name[@]}" ; do
    echov "* removing temporary file \"kubeconfig_${cp_name[i]}\""
    rm "kubeconfig_${cp_name[i]}" 2> /dev/null
done

if [[ "$arg_verbose" == "true"  && "${out_kubeconfig}" != "-" ]] ; then
    echov "Contexts of kubeconfig \"${out_kubeconfig}\":"
    kubectl --kubeconfig "${out_kubeconfig}" config get-contexts
fi
