package operator

import (
	"context"
	"time"

	operatorconfigclient "github.com/knelasevero/extended-job-manager-operator/pkg/generated/clientset/versioned"
	operatorclientinformers "github.com/knelasevero/extended-job-manager-operator/pkg/generated/informers/externalversions"
	"github.com/knelasevero/extended-job-manager-operator/pkg/operator/operatorclient"
	openshiftrouteclientset "github.com/openshift/client-go/route/clientset/versioned"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/operator/loglevel"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	workQueueKey          = "key"
	workQueueCMChangedKey = "CMkey"
)

type queueItem struct {
	kind string
	name string
}

func RunOperator(ctx context.Context, cc *controllercmd.ControllerContext) error {
	kubeClient, err := kubernetes.NewForConfig(cc.ProtoKubeConfig)
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(cc.ProtoKubeConfig)
	if err != nil {
		return err
	}

	kubeInformersForNamespaces := v1helpers.NewKubeInformersForNamespaces(kubeClient,
		"",
		operatorclient.OperatorNamespace,
	)

	operatorConfigClient, err := operatorconfigclient.NewForConfig(cc.KubeConfig)
	if err != nil {
		return err
	}
	operatorConfigInformers := operatorclientinformers.NewSharedInformerFactory(operatorConfigClient, 10*time.Minute)
	ejmClient := &operatorclient.ExtendedJobManagerClient{
		Ctx:            ctx,
		SharedInformer: operatorConfigInformers.Extendedjobmanager().V1().ExtendedJobManagers().Informer(),
		OperatorClient: operatorConfigClient.ExtendedjobmanagerV1(),
	}

	osrClient, err := openshiftrouteclientset.NewForConfig(cc.KubeConfig)
	if err != nil {
		return err
	}

	targetConfigReconciler, err := NewTargetConfigReconciler(
		ctx,
		operatorConfigClient.ExtendedjobmanagerV1(),
		operatorConfigInformers.Extendedjobmanager().V1().ExtendedJobManagers(),
		kubeInformersForNamespaces,
		ejmClient,
		kubeClient,
		osrClient,
		dynamicClient,
		cc.EventRecorder,
	)
	if err != nil {
		return err
	}

	logLevelController := loglevel.NewClusterOperatorLoggingController(ejmClient, cc.EventRecorder)

	klog.Infof("Starting informers")
	operatorConfigInformers.Start(ctx.Done())
	kubeInformersForNamespaces.Start(ctx.Done())

	klog.Infof("Starting log level controller")
	go logLevelController.Run(ctx, 1)
	klog.Infof("Starting target config reconciler")
	go targetConfigReconciler.Run(1, ctx.Done())

	<-ctx.Done()
	return nil

}
