package utilizeset

import (
	"context"
	"reflect"

	utilizev1alpha1 "kube-utilize-operator/pkg/apis/utilize/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var log = logf.Log.WithName("controller_utilizeset")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new UtilizeSet Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileUtilizeSet{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("utilizeset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource UtilizeSet
	err = c.Watch(&source.Kind{Type: &utilizev1alpha1.UtilizeSet{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner UtilizeSet
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &utilizev1alpha1.UtilizeSet{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileUtilizeSet implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileUtilizeSet{}

// ReconcileUtilizeSet reconciles a UtilizeSet object
type ReconcileUtilizeSet struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// This function will return all available resources within the cluster
func getAvailableResources(nodeNonTerminatedPodsList *corev1.PodList, nodeList *corev1.NodeList) {
	// get total resources
	var totalMemAvail int64
	var totalCPUAvail int64
	if len(nodeList.Items) > 0 {
		valid := true
		i := 0
		for valid {
			node := &nodeList.Items[i]

			// total available resources
			// TODO: Move this to separate func
			allocatable := node.Status.Allocatable
			memAvail := allocatable.Memory().Value()
			cpuAvail := allocatable.Cpu().MilliValue()
			// log.Info("Allocatable resources for Node", "Node", node, "Memory", (memAvail/1024)/1024, "CPU", cpuAvail)
			totalMemAvail = totalMemAvail + memAvail
			totalCPUAvail = totalCPUAvail + cpuAvail

			if i+1 == len(nodeList.Items) {
				valid = false
			}
			i++
		}
		log.Info("Total allocatable resources", "Memory", (totalMemAvail/1024)/1024, "CPU", totalCPUAvail)
	} else {
		log.Info("Unable to read nodes", "Nodes", len(nodeList.Items))
	}
	
	// get used resources
	log.Info("Non-terminated Pods","Size", len(nodeNonTerminatedPodsList.Items))
	
	// for each pod get total limit cpu & mem
	for _, pod := range nodeNonTerminatedPodsList.Items {
		// For each container get resource limits
		containerList := pod.Spec.Containers
		for _, container := range containerList {
			resources := container.Resources.Requests
			log.Info("Request resources for pod", "Pod labels", pod.ObjectMeta.Labels)
			log.Info("Container" ,"Container image", container.Image)
			log.Info("Resources", "Resource requests", resources)
		}
	}

	// calc available resources - used resources

}

// TODO: Get total usable resources

// Reconcile reads that state of the cluster for a UtilizeSet object and makes changes based on the state read
// and what is in the UtilizeSet.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileUtilizeSet) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling UtilizeSet")

	// Fetch the UtilizeSet instance
	utilizeSet := &utilizev1alpha1.UtilizeSet{}
	err := r.client.Get(context.TODO(), request.NamespacedName, utilizeSet)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// List all pods owned by this UtilizeSet instance
	lbls := labels.Set{
		"app":     utilizeSet.Name,
		"version": "v0.1",
	}
	existingPods := &corev1.PodList{}
	err = r.client.List(context.TODO(),
		existingPods,
		&client.ListOptions{
			Namespace:     request.Namespace,
			LabelSelector: labels.SelectorFromSet(lbls),
		})
	if err != nil {
		reqLogger.Error(err, "failed to list existing pods in the utilizeSet")
		return reconcile.Result{}, err
	}
	existingPodNames := []string{}
	// Count the pods that are pending or running as available
	for _, pod := range existingPods.Items {
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			continue
		}
		if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning {
			existingPodNames = append(existingPodNames, pod.GetObjectMeta().GetName())
		}
	}

	config, err := rest.InClusterConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
        reqLogger.Error(err, "Could not create clientset for kube api")
	}
	cluster := clientset.CoreV1()
	// 	retrieve allocatable cluster resources
	/*
		TODO: exclude pods from this operator
		TODO: Iterate over all nodes
		TODO: make sure unschedulable is false
		TODO: make all namespaces
	*/
	namespace := "default"
	nodeList, nodeListErr := cluster.Nodes().List(metav1.ListOptions{})
	nodeNonTerminatedPodsList, podListErr := cluster.Pods(namespace).List(metav1.ListOptions{})
	if nodeListErr == nil || podListErr == nil {
		reqLogger.Error(err, "Error while reading node & pod list data: %v")
	} else {
		getAvailableResources(nodeNonTerminatedPodsList, nodeList)
	}
	
	// TODO: calculate the total desired pods based on
	// resources / available resouces
	

	reqLogger.Info("Checking utilizeset", "expected replicas", utilizeSet.Spec.Replicas, "Pod.Names", existingPodNames)
	
	// Update the status if necessary
	status := utilizev1alpha1.UtilizeSetStatus{
		Replicas: int32(len(existingPodNames)),
		PodNames: existingPodNames,
	//	TotalMem: totalMemAvail,
	}
	if !reflect.DeepEqual(utilizeSet.Status, status) {
		utilizeSet.Status = status
		err := r.client.Status().Update(context.TODO(), utilizeSet)
		if err != nil {
			reqLogger.Error(err, "failed to update the utilizeSet")
			return reconcile.Result{}, err
		}
	}

	// Scale Down Pods
	if int32(len(existingPodNames)) > utilizeSet.Spec.Replicas {
		// delete a pod. Just one at a time (this reconciler will be called again afterwards)
		reqLogger.Info("Deleting a pod in the utilizeset", "expected replicas", utilizeSet.Spec.Replicas, "Pod.Names", existingPodNames)
		pod := existingPods.Items[0]
		err = r.client.Delete(context.TODO(), &pod)
		if err != nil {
			reqLogger.Error(err, "failed to delete a pod")
			return reconcile.Result{}, err
		}
	}

	// Scale Up Pods
	if int32(len(existingPodNames)) < utilizeSet.Spec.Replicas {
		// create a new pod. Just one at a time (this reconciler will be called again afterwards)
		reqLogger.Info("Adding a pod in the utilizeset", "expected replicas", utilizeSet.Spec.Replicas, "Pod.Names", existingPodNames)
		pod := newPodForCR(utilizeSet)
		if err := controllerutil.SetControllerReference(utilizeSet, pod, r.scheme); err != nil {
			reqLogger.Error(err, "unable to set owner reference on new pod")
			return reconcile.Result{}, err
		}
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			reqLogger.Error(err, "failed to create a pod")
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{Requeue: true}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *utilizev1alpha1.UtilizeSet) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
		"version": "v0.1",
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:	cr.Name + "-pod",
			Namespace: 		cr.Namespace,
			Labels:    		labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
