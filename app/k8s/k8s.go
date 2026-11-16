package k8s

import (
	"context"
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func NewClient() (client *kubernetes.Clientset, err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return
		}
	}
	client, err = kubernetes.NewForConfig(config)
	return
}

func NewServiceInformer(client *kubernetes.Clientset, period time.Duration) cache.SharedIndexInformer {
	filter := fields.OneTermEqualSelector("spec.nodeName", os.Getenv("NODE_NAME")).String()

	factory := informers.NewSharedInformerFactoryWithOptions(client, period, informers.WithTweakListOptions(
		func(lo *metav1.ListOptions) {
			lo.FieldSelector = filter
		},
	))

	informer := factory.Core().V1().Services().Informer()
	return informer // just care about services
}

func NewPodInformer(client *kubernetes.Clientset, period time.Duration) cache.SharedIndexInformer {
	filter := fields.OneTermEqualSelector("spec.nodeName", os.Getenv("NODE_NAME")).String()

	factory := informers.NewSharedInformerFactoryWithOptions(client, period, informers.WithTweakListOptions(
		func(lo *metav1.ListOptions) {
			lo.FieldSelector = filter
		},
	))

	informer := factory.Core().V1().Pods().Informer()
	return informer // just care about pods
}
