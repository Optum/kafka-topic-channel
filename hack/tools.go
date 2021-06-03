// +build tools

package tools

import (
	_ "knative.dev/hack"
	_ "knative.dev/pkg/hack"

	_ "knative.dev/reconciler-test/cmd/eventshub"
)
