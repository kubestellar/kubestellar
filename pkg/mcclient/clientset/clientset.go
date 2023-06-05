package clientset

import (
	discovery "k8s.io/client-go/discovery"
	kubeclient "k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"

	ksclient "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned"
)

type Interface interface {
	//	Discovery() discovery.DiscoveryInterface
	// Kube returns Kubernetes client interface
	Kube() kubeclient.Interface
	// KS returns Kubestellar client interface
	KS() ksclient.Interface
}

// Clientset contains the clients for groups.
type Clientset struct {
	*discovery.DiscoveryClient
	kube        *kubeclient.Clientset
	kubestellar *ksclient.Clientset
}

func (c *Clientset) Kube() kubeclient.Interface {
	return c.kube
}

func (c *Clientset) KS() ksclient.Interface {
	return c.kubestellar
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*Clientset, error) {

	var cs Clientset
	var err error
	cs.kube, err = kubeclient.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	cs.kubestellar, err = ksclient.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return &cs, nil

}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	cs, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.kube = kubeclient.New(c)
	cs.kubestellar = ksclient.New(c)
	return &cs
}
