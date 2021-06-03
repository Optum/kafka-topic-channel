package defaultsystem

import (
	"os"

	"knative.dev/pkg/system"
)

func init() {
	if ns := os.Getenv(system.NamespaceEnvKey); ns != "" {
		return
	}
	os.Setenv(system.NamespaceEnvKey, "knative-eventing")
}
