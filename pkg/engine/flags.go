package engine

import (
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	ClamavAgentImage      string
	ClamavUpdaterImage    string
	ClamavLibraryHostPath string

	ClamavUpdaterCPURequest    = ResourceQuantityFlag{Quantity: resource.MustParse("50m")}
	ClamavUpdaterCPULimit      = ResourceQuantityFlag{Quantity: resource.MustParse("200m")}
	ClamavUpdaterMemoryRequest = ResourceQuantityFlag{Quantity: resource.MustParse("192Mi")}
	ClamavUpdaterMemoryLimit   = ResourceQuantityFlag{Quantity: resource.MustParse("192Mi")}
)

func FlagSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("engine", pflag.ExitOnError)
	set.StringVar(&ClamavAgentImage, "engine-clamav-agent-image", "quay.io/mittwald/kubeav-agent-clamav:v1", "image to use for ClamAV scans")
	set.StringVar(&ClamavUpdaterImage, "engine-clamav-updater-image", "quay.io/mittwald/kubeav-updater-clamav:v1", "image to use for ClamAV updater")
	set.StringVar(&ClamavLibraryHostPath, "engine-clamav-library", "/var/lib/clamav", "path to ClamAV library on node")

	set.Var(&ClamavUpdaterCPURequest, "engine-clamav-updater-cpu-request", "CPU requests for ClamAV updater")
	set.Var(&ClamavUpdaterCPULimit, "engine-clamav-updater-cpu-limit", "CPU limits for ClamAV updater")
	set.Var(&ClamavUpdaterMemoryRequest, "engine-clamav-updater-memory-request", "Memory requests for ClamAV updater")
	set.Var(&ClamavUpdaterMemoryLimit, "engine-clamav-updater-memory-limit", "Memory limits for ClamAV updater")

	return set
}

type ResourceQuantityFlag struct {
	resource.Quantity
}

func (r *ResourceQuantityFlag) String() string {
	return r.Quantity.String()
}

func (r *ResourceQuantityFlag) Set(s string) error {
	parsed, err := resource.ParseQuantity(s)
	if err != nil {
		return err
	}

	r.Quantity = parsed
	return nil
}

func (r *ResourceQuantityFlag) Type() string {
	return "resources"
}
