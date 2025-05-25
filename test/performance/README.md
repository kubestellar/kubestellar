## Workload Benchmark for KubeStellar

### Pre-requisite: 

In order to follow the instructions below, you must have [python3](https://www.python.org/downloads/) with all the dependencies listed [here](common/requirements.txt) installed. We recommend to create a python virtual environment `.venv` under `test/performance/common/`, for example: 

```bash
cd test/performance/common 
python3 -m venv .venv
. .venv/bin/activate
pip3 install -r requirements.txt
```

Additionally, you must have an environment with KubeStellar installed; see [KubeStellar getting started](https://docs.kubestellar.io/release-0.23.1/direct/get-started/). Alternatively, you can also use KubeStellar e2e script [run-test.sh](https://github.com/kubestellar/kubestellar/blob/main/test/e2e/run-test.sh) to setup an environment.

Use the following instructions to generate the sample workload for KubeStellar performance experiments:

a)[Instructions for short-running tests](short-running-tests/README.md)

b)[Instructions for long-running tests](long-running-tests/README.md)
