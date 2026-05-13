package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	challengev1 "github.com/rodrigomicrosiga/k8s-operator-challenge/api/v1"
)

// OperatorChallengeReconciler reconciles a OperatorChallenge object
type OperatorChallengeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.rodrigo.com,resources=operatorchallenges,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.rodrigo.com,resources=operatorchallenges/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.rodrigo.com,resources=operatorchallenges/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *OperatorChallengeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. Buscar o CR OperatorChallenge
	var cr challengev1.OperatorChallenge
	if err := r.Get(ctx, req.NamespacedName, &cr); err != nil {
		// recurso pode ter sido deletado
		logger.Error(err, "unable to fetch OperatorChallenge", "namespacedName", req.NamespacedName)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Construir o Deployment desejado e garantir owner reference
	desiredDep := r.deploymentFor(&cr)
	if err := controllerutil.SetControllerReference(&cr, desiredDep, r.Scheme); err != nil {
		logger.Error(err, "unable to set owner reference on Deployment")
		return ctrl.Result{}, err
	}

	var existingDep appsv1.Deployment
	depKey := types.NamespacedName{Name: desiredDep.Name, Namespace: desiredDep.Namespace}
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, &existingDep, func() error {
		// manter labels e spec do desiredDep
		existingDep.Labels = desiredDep.Labels
		existingDep.Spec = desiredDep.Spec
		return nil
	})
	if err != nil {
		logger.Error(err, "failed to create or update Deployment", "deployment", depKey)
		return ctrl.Result{}, err
	}
	logger.Info("Deployment reconciled", "operation", op, "deployment", depKey)

	// 3. Construir o Service desejado e garantir owner reference
	desiredSvc := r.serviceFor(&cr)
	if err := controllerutil.SetControllerReference(&cr, desiredSvc, r.Scheme); err != nil {
		logger.Error(err, "unable to set owner reference on Service")
		return ctrl.Result{}, err
	}

	var existingSvc corev1.Service
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, &existingSvc, func() error {
		existingSvc.Labels = desiredSvc.Labels
		existingSvc.Spec = desiredSvc.Spec
		return nil
	})
	if err != nil {
		logger.Error(err, "failed to create or update Service", "service", types.NamespacedName{Name: desiredSvc.Name, Namespace: desiredSvc.Namespace})
		return ctrl.Result{}, err
	}
	logger.Info("Service reconciled", "service", types.NamespacedName{Name: desiredSvc.Name, Namespace: desiredSvc.Namespace})

	// 4. Atualizar status com readyReplicas do Deployment
	var updatedDep appsv1.Deployment
	if err := r.Get(ctx, depKey, &updatedDep); err == nil {
		ready := updatedDep.Status.ReadyReplicas
		if cr.Status.ReadyReplicas != ready {
			cr.Status.ReadyReplicas = ready
			if err := r.Status().Update(ctx, &cr); err != nil {
				logger.Error(err, "failed to update OperatorChallenge status")
				return ctrl.Result{}, err
			}
			logger.Info("OperatorChallenge status updated", "readyReplicas", ready)
		}
	} else {
		// se não conseguiu obter o deployment, loga e continua (CreateOrUpdate já tratou criação)
		logger.Error(err, "failed to get Deployment for status update", "deployment", depKey)
	}

	return ctrl.Result{}, nil
}

func (r *OperatorChallengeReconciler) deploymentFor(cr *challengev1.OperatorChallenge) *appsv1.Deployment {
	labels := map[string]string{"app": cr.Name}

	replicas := int32(1)
	if cr.Spec.Replicas != nil {
		replicas = *cr.Spec.Replicas
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-deployment",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: cr.Spec.Image,
							Ports: []corev1.ContainerPort{
								{ContainerPort: cr.Spec.Port},
							},
						},
					},
				},
			},
		},
	}
}

func (r *OperatorChallengeReconciler) serviceFor(cr *challengev1.OperatorChallenge) *corev1.Service {
	labels := map[string]string{"app": cr.Name}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-svc",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{Port: cr.Spec.Port},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func (r *OperatorChallengeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&challengev1.OperatorChallenge{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
