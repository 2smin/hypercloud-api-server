package caller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	clusterv1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/cluster/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var remoteClientset *kubernetes.Clientset
var remoteRestConfig *restclient.Config

func CreateRoleInRemote(clusterManager *clusterv1alpha1.ClusterManager, subject string, remoteRole string, attribute string) error {
	if remoteRole == "admin" {
		remoteRole = "cluster-admin"
	}
	remoteClientset, err := getRemoteK8sClient(clusterManager)
	if err != nil {
		return err
	}

	// var clusterRoleName string
	var clusterRoleBindingName string
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if attribute == "user" {
		// clusterRoleName = subject + "-user-" + clusterManager.Name + "-clm-role"
		clusterRoleBindingName = subject + "-user-rolebinding"
		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     subject,
			},
		}
	} else {
		// clusterRoleName = subject + "-group-" + clusterManager.Name + "-clm-role"
		clusterRoleBindingName = subject + "-group-rolebinding"
		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Group",
				Name:     subject,
			},
		}
	}

	clusterRoleBinding.ObjectMeta = metav1.ObjectMeta{
		Name: clusterRoleBindingName,
	}
	clusterRoleBinding.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     remoteRole,
	}

	if _, err := remoteClientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{}); err != nil {
		klog.V(1).Infoln(err)
		return err
	}
	msg := "Create clusterrole [" + remoteRole + "] to remote cluster [" + clusterManager.Name + "] for subject [" + subject + "] "
	klog.V(3).Infoln(msg)
	return nil
}

func RemoveRoleFromRemote(clusterManager *clusterv1alpha1.ClusterManager, subject string, attribute string) error {
	remoteClientset, err := getRemoteK8sClient(clusterManager)
	if err != nil {
		return err
	}

	// var clusterRoleName string
	var clusterRoleBindingName string
	if attribute == "user" {
		clusterRoleBindingName = subject + "-user-rolebinding"
	} else {
		clusterRoleBindingName = subject + "-group-rolebinding"
	}

	if _, err := remoteClientset.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			klog.V(3).Infoln("Rolebinding [" + clusterRoleBindingName + "] is already deleted")
			return nil
		} else {
			klog.V(1).Infoln(err)
			return err
		}
	} else {
		if err := remoteClientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), clusterRoleBindingName, metav1.DeleteOptions{}); err != nil {
			klog.V(1).Infoln(err)
			return err
		}
	}

	msg := "Remove rolebinding [" + clusterRoleBindingName + "] from remote cluster [" + clusterManager.Name + "] for subject [" + subject + "]"
	klog.V(3).Infoln(msg)
	return nil
}

func getRemoteK8sClient(clusterManager *clusterv1alpha1.ClusterManager) (*kubernetes.Clientset, error) {
	if remoteKubeconfig, err := Clientset.CoreV1().Secrets(clusterManager.Namespace).Get(context.TODO(), clusterManager.Name+"-kubeconfig", metav1.GetOptions{}); err == nil {
		if value, ok := remoteKubeconfig.Data["value"]; ok {
			remoteClientConfig, err := clientcmd.NewClientConfigFromBytes(value)
			if err != nil {
				klog.V(1).Infoln(err)
				return nil, err
			}
			remoteRestConfig, err = remoteClientConfig.ClientConfig()
			if err != nil {
				klog.V(1).Infoln(err)
				return nil, err
			}
		}
		remoteClientset, err = kubernetes.NewForConfig(remoteRestConfig)
		if err != nil {
			klog.V(1).Infoln(err)
			return nil, err
		}
		return remoteClientset, nil
	} else if errors.IsNotFound(err) {
		klog.V(3).Infoln("Cluster [" + clusterManager.Name + "] is not ready yet")
		return nil, err
	} else {
		klog.V(1).Infoln("Error: Get clusterrole [" + clusterManager.Name + "] is failed")
		return nil, err
	}
}
