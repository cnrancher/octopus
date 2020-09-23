// +build test

package envtest

import (
	"github.com/rancher/octopus/pkg/util/log/zap"
)

var log = zap.WrapAsLogr(zap.NewDevelopmentLogger()).WithName("test-env")
