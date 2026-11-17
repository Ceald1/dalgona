package k8s

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	wg "github.com/Ceald1/dalgona/app/wireguard"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var NS = "dalgona-secrets"
var SecretNameBase = "pub-key"

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

type IPsAndServices map[string]string

// per node function, spits out IPsAndServices type (string map of strings)
func GetIPs(client *kubernetes.Clientset) (results IPsAndServices, err error) {
	filter := fields.OneTermEqualSelector("spec.nodeName", os.Getenv("NODE_NAME")).String()
	pods, err := client.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
		FieldSelector: filter,
	})
	if err != nil {
		return
	}
	results = make(IPsAndServices)
	for _, pod := range pods.Items {
		ip := pod.Status.PodIP
		results[fmt.Sprintf("%s.%s.cluster.local", strings.ReplaceAll(ip, ".", "-"), pod.Namespace)] = ip
	}
	services, err := client.CoreV1().Services("").List(context.Background(), metav1.ListOptions{
		FieldSelector: filter,
	})
	if err != nil {
		return
	}
	for _, service := range services.Items {
		results[fmt.Sprintf("%s.%s.cluster.local", strings.ReplaceAll(service.Name, "_", "-"), service.Namespace)] = service.Spec.ClusterIP
	}
	return results, nil
}

// update secret for node's public key, run after 5 minutes and have nodes ONLY update their key and no others
func UpdateSecret(client *kubernetes.Clientset, cfg wg.Conf) (err error) {
	secret, err := client.CoreV1().Secrets(NS).Get(Context.Background(), fmt.Sprintf("%s_%s", SecretName, os.Getenv("NODE_NAME")), metav1.GetOptions{})

}
