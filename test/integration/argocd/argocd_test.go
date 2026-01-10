package argocd

import (
	"os/exec"
	"testing"
	"time"
)

func TestArgoCDServerIsAvailable(t *testing.T) {
	t.Log("Waiting for Argo CD server deployment to become available")

	cmd := exec.Command(
		"kubectl",
		"wait",
		"--namespace", "argocd",
		"--for=condition=Available",
		"deployment/argocd-server",
		"--timeout=300s",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf(
			"Argo CD server did not become available: %v\nOutput:\n%s",
			err,
			string(output),
		)
	}

	t.Log("Argo CD server deployment is available")
	time.Sleep(1 * time.Second)
}
