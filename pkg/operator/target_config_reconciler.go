package operator

import (
	"context"
	"fmt"
	"strconv"
	"time"

	extendedjobmanagersv1 "github.com/knelasevero/extended-job-manager-operator/pkg/apis/extendedjobmanager/v1"
	operatorconfigclientv1 "github.com/knelasevero/extended-job-manager-operator/pkg/generated/clientset/versioned/typed/extendedjobmanager/v1"
	operatorclientinformers "github.com/knelasevero/extended-job-manager-operator/pkg/generated/informers/externalversions/extendedjobmanager/v1"
	"github.com/knelasevero/extended-job-manager-operator/pkg/operator/operatorclient"
	openshiftrouteclientset "github.com/openshift/client-go/route/clientset/versioned"
	"github.com/openshift/library-go/pkg/controller"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	operatorv1 "github.com/openshift/api/operator/v1"

	"github.com/knelasevero/extended-job-manager-operator/bindata"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	DefaultImage    = "gcr.io/k8s-staging-kueue/kueue"
	PromNamespace   = "openshift-monitoring"
	PromRouteName   = "prometheus-k8s"
	PromTokenPrefix = "prometheus-k8s-token"
)

var extendedJobManagerConfigMap = "extended-job-manager"

type TargetConfigReconciler struct {
	ctx                        context.Context
	operatorClient             operatorconfigclientv1.ExtendedjobmanagerV1Interface
	extendedJobManagerClient   *operatorclient.ExtendedJobManagerClient
	kubeClient                 kubernetes.Interface
	osrClient                  openshiftrouteclientset.Interface
	dynamicClient              dynamic.Interface
	eventRecorder              events.Recorder
	queue                      workqueue.RateLimitingInterface
	kubeInformersForNamespaces v1helpers.KubeInformersForNamespaces
}

func NewTargetConfigReconciler(
	ctx context.Context,
	operatorConfigClient operatorconfigclientv1.ExtendedjobmanagerV1Interface,
	operatorClientInformer operatorclientinformers.ExtendedJobManagerInformer,
	kubeInformersForNamespaces v1helpers.KubeInformersForNamespaces,
	extendedJobManagerClient *operatorclient.ExtendedJobManagerClient,
	kubeClient kubernetes.Interface,
	osrClient openshiftrouteclientset.Interface,
	dynamicClient dynamic.Interface,
	eventRecorder events.Recorder,
) (*TargetConfigReconciler, error) {
	c := &TargetConfigReconciler{
		ctx:                        ctx,
		operatorClient:             operatorConfigClient,
		extendedJobManagerClient:   extendedJobManagerClient,
		kubeClient:                 kubeClient,
		osrClient:                  osrClient,
		dynamicClient:              dynamicClient,
		eventRecorder:              eventRecorder,
		queue:                      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "TargetConfigReconciler"),
		kubeInformersForNamespaces: kubeInformersForNamespaces,
	}

	_, err := operatorClientInformer.Informer().AddEventHandler(c.eventHandler(queueItem{kind: "extendedjobmanager"}))
	if err != nil {
		return nil, err
	}

	_, err = kubeInformersForNamespaces.InformersFor(operatorclient.OperatorNamespace).Core().V1().ConfigMaps().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {},
		UpdateFunc: func(old, new interface{}) {
			cm, ok := old.(*v1.ConfigMap)
			if !ok {
				klog.Errorf("Unable to convert obj to ConfigMap")
				return
			}
			c.queue.Add(queueItem{kind: "configmap", name: cm.Name})
		},
		DeleteFunc: func(obj interface{}) {
			cm, ok := obj.(*v1.ConfigMap)
			if !ok {
				klog.Errorf("Unable to convert obj to ConfigMap")
				return
			}
			c.queue.Add(queueItem{kind: "configmap", name: cm.Name})
		},
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c TargetConfigReconciler) sync(item queueItem) error {
	extendedJobManager, err := c.operatorClient.ExtendedJobManagers(operatorclient.OperatorNamespace).Get(c.ctx, operatorclient.OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "unable to get operator configuration", "namespace", operatorclient.OperatorNamespace, "extended-job-manager", operatorclient.OperatorConfigName)
		return err
	}

	specAnnotations := map[string]string{
		"extendedjobmanagers.operator.openshift.io/cluster": strconv.FormatInt(extendedJobManager.Generation, 10),
	}

	// Skip any sync triggered by other than the ExtendedJobManagerConfig CM changes
	if item.kind == "configmap" {
		if item.name != extendedJobManager.Spec.ExtendedJobManagerConfig{
			return nil
		}
		klog.Infof("configmap %q changed, forcing redeployment", extendedJobManager.Spec.ExtendedJobManagerConfig)
	}

	configMapResourceVersion, err := c.getConfigMapResourceVersion(extendedJobManager)
	if err != nil {
		return err
	}
	specAnnotations["configmaps/"+extendedJobManager.Spec.ExtendedJobManagerConfig] = configMapResourceVersion

	if sa, _, err := c.manageServiceAccount(extendedJobManager); err != nil {
		return err
	} else {
		resourceVersion := "0"
		if sa != nil { // SyncConfigMap can return nil
			resourceVersion = sa.ObjectMeta.ResourceVersion
		}
		specAnnotations["serviceaccounts/extended-job-manager"] = resourceVersion
	}

	if clusterRole, _, err := c.manageClusterRole(extendedJobManager); err != nil {
		return err
	} else {
		resourceVersion := "0"
		if clusterRole != nil { // SyncConfigMap can return nil
			resourceVersion = clusterRole.ObjectMeta.ResourceVersion
		}
		specAnnotations["clusterroles/extended-job-manager"] = resourceVersion
	}

	if clusterRoleBindings, _, err := c.manageClusterRoleBindings(extendedJobManager); err != nil {
		return err
	} else {
		resourceVersion := "0"
		if clusterRoleBindings != nil { // SyncConfigMap can return nil
			resourceVersion = clusterRoleBindings.ObjectMeta.ResourceVersion
		}
		specAnnotations["clusterrolebindings/extended-job-manager"] = resourceVersion
	}

	if role, _, err := c.manageRole(extendedJobManager); err != nil {
		return err
	} else {
		resourceVersion := "0"
		if role != nil { // SyncConfigMap can return nil
			resourceVersion = role.ObjectMeta.ResourceVersion
		}
		specAnnotations["roles/extended-job-manager"] = resourceVersion
	}

	if roleBinding, _, err := c.manageRoleBinding(extendedJobManager); err != nil {
		return err
	} else {
		resourceVersion := "0"
		if roleBinding != nil { // SyncConfigMap can return nil
			resourceVersion = roleBinding.ObjectMeta.ResourceVersion
		}
		specAnnotations["rolebindings/extended-job-manager"] = resourceVersion
	}

	deployment, _, err := c.manageDeployment(extendedJobManager, specAnnotations)
	if err != nil {
		return err
	}

	_, _, err = v1helpers.UpdateStatus(c.ctx, c.extendedJobManagerClient, func(status *operatorv1.OperatorStatus) error {
		resourcemerge.SetDeploymentGeneration(&status.Generations, deployment)
		return nil
	})
	return err
}

func (c *TargetConfigReconciler) manageConfigMap(extendedJobManager *extendedjobmanagersv1.ExtendedJobManager) (*v1.ConfigMap, bool, error) {
	var required *v1.ConfigMap
	var err error

	required, err = c.kubeClient.CoreV1().ConfigMaps(extendedJobManager.Namespace).Get(context.TODO(), string(extendedJobManager.Spec.ExtendedJobManagerConfig), metav1.GetOptions{})

	if err != nil {
		klog.Errorf("Cannot load ConfigMap %s for the extendedjobmanager", string(extendedJobManager.Spec.ExtendedJobManagerConfig))
		return nil, false, err
	}

	extendedJobManagerConfigMap = string(extendedJobManager.Spec.ExtendedJobManagerConfig)
	klog.Infof("Find ConfigMap %s for the extendedjobmanager.", extendedJobManager.Spec.ExtendedJobManagerConfig)

	return resourceapply.ApplyConfigMap(c.ctx, c.kubeClient.CoreV1(), c.eventRecorder, required)
}

func (c *TargetConfigReconciler) getConfigMapResourceVersion(extendedJobManager *extendedjobmanagersv1.ExtendedJobManager) (string, error) {
	required, err := c.kubeInformersForNamespaces.InformersFor(operatorclient.OperatorNamespace).Core().V1().ConfigMaps().Lister().ConfigMaps(operatorclient.OperatorNamespace).Get(extendedJobManager.Spec.ExtendedJobManagerConfig)
	if err != nil {
		return "", fmt.Errorf("could not get configuration configmap: %v", err)
	}

	return required.ObjectMeta.ResourceVersion, nil
}

func (c *TargetConfigReconciler) manageServiceAccount(extendedJobManager *extendedjobmanagersv1.ExtendedJobManager) (*v1.ServiceAccount, bool, error) {
	required := resourceread.ReadServiceAccountV1OrDie(bindata.MustAsset("assets/extended-job-manager/serviceaccount.yaml"))
	required.Namespace = extendedJobManager.Namespace
	ownerReference := metav1.OwnerReference{
		APIVersion: "operator.openshift.io/v1",
		Kind:       "ExtendedJobManager",
		Name:       extendedJobManager.Name,
		UID:        extendedJobManager.UID,
	}
	required.OwnerReferences = []metav1.OwnerReference{
		ownerReference,
	}
	controller.EnsureOwnerRef(required, ownerReference)

	return resourceapply.ApplyServiceAccount(c.ctx, c.kubeClient.CoreV1(), c.eventRecorder, required)
}

func (c *TargetConfigReconciler) manageClusterRole(extendedJobManager *extendedjobmanagersv1.ExtendedJobManager) (*rbacv1.ClusterRole, bool, error) {
	required := resourceread.ReadClusterRoleV1OrDie(bindata.MustAsset("assets/extended-job-manager/clusterrole-extended-job-manager.yaml"))
	ownerReference := metav1.OwnerReference{
		APIVersion: "operator.openshift.io/v1",
		Kind:       "ExtendedJobManager",
		Name:       extendedJobManager.Name,
		UID:        extendedJobManager.UID,
	}
	required.OwnerReferences = []metav1.OwnerReference{
		ownerReference,
	}
	controller.EnsureOwnerRef(required, ownerReference)

	return resourceapply.ApplyClusterRole(c.ctx, c.kubeClient.RbacV1(), c.eventRecorder, required)
}

func (c *TargetConfigReconciler) manageClusterRoleBindings(extendedJobManager *extendedjobmanagersv1.ExtendedJobManager) (*rbacv1.ClusterRoleBinding, bool, error) {
	required := resourceread.ReadClusterRoleBindingV1OrDie(bindata.MustAsset("assets/extended-job-manager/clusterrolebinding-extended-job-manager.yaml"))
	ownerReference := metav1.OwnerReference{
		APIVersion: "operator.openshift.io/v1",
		Kind:       "ExtendedJobManager",
		Name:       extendedJobManager.Name,
		UID:        extendedJobManager.UID,
	}
	required.OwnerReferences = []metav1.OwnerReference{
		ownerReference,
	}
	controller.EnsureOwnerRef(required, ownerReference)

	return resourceapply.ApplyClusterRoleBinding(c.ctx, c.kubeClient.RbacV1(), c.eventRecorder, required)
}

func (c *TargetConfigReconciler) manageRole(extendedJobManager *extendedjobmanagersv1.ExtendedJobManager) (*rbacv1.Role, bool, error) {
	required := resourceread.ReadRoleV1OrDie(bindata.MustAsset("assets/extended-job-manager/role-extended-job-manager.yaml"))
	required.Namespace = extendedJobManager.Namespace
	ownerReference := metav1.OwnerReference{
		APIVersion: "operator.openshift.io/v1",
		Kind:       "ExtendedJobManager",
		Name:       extendedJobManager.Name,
		UID:        extendedJobManager.UID,
	}
	required.OwnerReferences = []metav1.OwnerReference{
		ownerReference,
	}
	controller.EnsureOwnerRef(required, ownerReference)

	return resourceapply.ApplyRole(c.ctx, c.kubeClient.RbacV1(), c.eventRecorder, required)
}

func (c *TargetConfigReconciler) manageRoleBinding(extendedJobManager *extendedjobmanagersv1.ExtendedJobManager) (*rbacv1.RoleBinding, bool, error) {
	required := resourceread.ReadRoleBindingV1OrDie(bindata.MustAsset("assets/extended-job-manager/rolebinding-extended-job-manager.yaml"))
	required.Namespace = extendedJobManager.Namespace
	ownerReference := metav1.OwnerReference{
		APIVersion: "operator.openshift.io/v1",
		Kind:       "ExtendedJobManager",
		Name:       extendedJobManager.Name,
		UID:        extendedJobManager.UID,
	}
	required.OwnerReferences = []metav1.OwnerReference{
		ownerReference,
	}
	controller.EnsureOwnerRef(required, ownerReference)

	return resourceapply.ApplyRoleBinding(c.ctx, c.kubeClient.RbacV1(), c.eventRecorder, required)
}

func (c *TargetConfigReconciler) manageDeployment(extendedJobManager *extendedjobmanagersv1.ExtendedJobManager, specAnnotations map[string]string) (*appsv1.Deployment, bool, error) {
	required := resourceread.ReadDeploymentV1OrDie(bindata.MustAsset("assets/extended-job-manager/deployment.yaml"))
	required.Name = operatorclient.OperandName
	required.Namespace = extendedJobManager.Namespace
	ownerReference := metav1.OwnerReference{
		APIVersion: "operator.openshift.io/v1",
		Kind:       "ExtendedJobManager",
		Name:       extendedJobManager.Name,
		UID:        extendedJobManager.UID,
	}
	required.OwnerReferences = []metav1.OwnerReference{
		ownerReference,
	}
	controller.EnsureOwnerRef(required, ownerReference)

	images := map[string]string{
		"${IMAGE}": extendedJobManager.Spec.ExtendedJobManagerImage,
	}
	for i := range required.Spec.Template.Spec.Containers {
		for pat, img := range images {
			if required.Spec.Template.Spec.Containers[i].Image == pat {
				required.Spec.Template.Spec.Containers[i].Image = img
				break
			}
		}
	}

	configmaps := map[string]string{
		"${CONFIGMAP}": extendedJobManager.Spec.ExtendedJobManagerConfig,
	}

	for i := range required.Spec.Template.Spec.Volumes {
		for pat, configmap := range configmaps {
			if required.Spec.Template.Spec.Volumes[i].ConfigMap.Name == pat {
				required.Spec.Template.Spec.Volumes[i].ConfigMap.Name = configmap
				break
			}
		}
	}

	// kueue not supporting -v
	// switch extendedJobManager.Spec.LogLevel {
	// case operatorv1.Normal:
	// 	required.Spec.Template.Spec.Containers[0].Args = append(required.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("-v=%d", 2))
	// case operatorv1.Debug:
	// 	required.Spec.Template.Spec.Containers[0].Args = append(required.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("-v=%d", 4))
	// case operatorv1.Trace:
	// 	required.Spec.Template.Spec.Containers[0].Args = append(required.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("-v=%d", 6))
	// case operatorv1.TraceAll:
	// 	required.Spec.Template.Spec.Containers[0].Args = append(required.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("-v=%d", 8))
	// default:
	// 	required.Spec.Template.Spec.Containers[0].Args = append(required.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("-v=%d", 2))
	// }

	resourcemerge.MergeMap(resourcemerge.BoolPtr(false), &required.Spec.Template.Annotations, specAnnotations)

	return resourceapply.ApplyDeployment(
		c.ctx,
		c.kubeClient.AppsV1(),
		c.eventRecorder,
		required,
		resourcemerge.ExpectedDeploymentGeneration(required, extendedJobManager.Status.Generations))
}

// Run starts the runner and blocks until stopCh is closed.
func (c *TargetConfigReconciler) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Infof("Starting TargetConfigReconciler")
	defer klog.Infof("Shutting down TargetConfigReconciler")

	// doesn't matter what workers say, only start one.
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
}

func (c *TargetConfigReconciler) processNextWorkItem() bool {
	dsKey, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(dsKey)
	item := dsKey.(queueItem)
	err := c.sync(item)
	if err == nil {
		c.queue.Forget(dsKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
	c.queue.AddRateLimited(dsKey)

	return true
}

func (c *TargetConfigReconciler) runWorker() {
	for c.processNextWorkItem() {
	}
}

// eventHandler queues the operator to check spec and status
func (c *TargetConfigReconciler) eventHandler(item queueItem) cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.queue.Add(item) },
		UpdateFunc: func(old, new interface{}) { c.queue.Add(item) },
		DeleteFunc: func(obj interface{}) { c.queue.Add(item) },
	}
}
