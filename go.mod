module github.com/coolsun/cloud-app

go 1.15

require (
	github.com/caarlos0/env/v6 v6.5.0
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/sessions v0.0.3
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator/v10 v10.4.1
	github.com/go-sql-driver/mysql v1.5.0
	github.com/imroc/req v0.3.0
	github.com/json-iterator/go v1.1.11
	github.com/kubernetes-csi/external-snapshotter/client/v4 v4.1.0
	github.com/mittwald/go-helm-client v0.5.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.3.0
	github.com/swaggo/swag v1.7.0
	github.com/thinkerou/favicon v0.1.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	helm.sh/helm/v3 v3.5.1
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.20.2
	k8s.io/code-generator v0.20.1
	k8s.io/metrics v0.20.1
	sigs.k8s.io/controller-runtime v0.8.1
	xorm.io/xorm v1.1.2
)

replace sigs.k8s.io/kustomize => ./utils/github/kustomize
