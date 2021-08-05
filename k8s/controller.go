package k8s

import (
	"context"
	cloudv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/utils/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	Kube *Kube
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cloud.k8s.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.k8s.io,resources=apps/status,verbs=get;update;patch

//todo 需要添加 finalizer 功能,使用 Finalizer 来做资源的清理
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	app := &cloudv1.App{}
	err := r.Get(context.TODO(), req.NamespacedName, app)
	if err != nil {
		if apierrors.IsNotFound(err) {
			//删除操作
			r.Kube.deleteApp(req.Namespace, req.Name)
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	} else {
		err = r.Kube.createApp(app)
		if err != nil {
			log.Errorf("%+v", err)
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudv1.App{}).
		Complete(r)
}

var (
	myScheme = runtime.NewScheme()
)

func controllerStart(kube *Kube) {
	_ = clientgoscheme.AddToScheme(myScheme)
	_ = cloudv1.AddToScheme(myScheme)
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme: myScheme,
	})
	if err != nil {
		panic(err)
	}
	appReconciler := &AppReconciler{
		Kube:   kube,
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	if err = appReconciler.SetupWithManager(mgr); err != nil {
		panic(err)
	}
	log.Info("crd控制器启动成功")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		panic(err)
	}
}
