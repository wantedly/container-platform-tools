package k8splatforms

import (
	"github.com/wantedly/container-platform-tools/dockerplatforms"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Row struct {
	Namespace         string
	ApiVersion        string
	Kind              string
	Name              string
	SubName           string
	ScheduledPlatform *dockerplatforms.DockerPlatform
	DeclaredPlatforms dockerplatforms.DockerPlatformList
	ImagePlatforms    dockerplatforms.DockerPlatformList
	HasViolation      bool
}

func EvaluateObject(
	obj client.Object,
	processors []KindProcessor,
) {
}
