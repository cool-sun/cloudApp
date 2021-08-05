package helmclient

import (
	"context"

	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
)

//go:generate mockgen -source=interface.go -package mockhelmclient -destination=./mock/interface.go -self_package=. Client

type Client interface {
	AddOrUpdateChartRepo(entry repo.Entry) error
	UpdateChartRepos() error
	InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) error
	ListDeployedReleases() ([]*release.Release, error)
	GetRelease(name string) (*release.Release, error)
	GetReleaseValues(name string, allValues bool) (map[string]interface{}, error)
	DeleteChartFromCache(spec *ChartSpec) error
	UninstallRelease(spec *ChartSpec) error
	TemplateChart(spec *ChartSpec) ([]byte, error)
	LintChart(spec *ChartSpec) error
}
