package k8s_d

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	wg "github.com/Ceald1/dalgona/app/wireguard"
	corev1 "k8s.io/api/core/v1"
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
	SecretName := fmt.Sprintf("%s-%s", SecretNameBase, os.Getenv("NODE_NAME"))
	secret, err := client.CoreV1().Secrets(NS).Get(context.Background(), SecretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	secret.Data["pubKey"] = []byte(cfg.PrivateKey.PublicKey().String())
	_, err = client.CoreV1().Secrets(NS).Update(context.Background(), secret, metav1.UpdateOptions{})

	return err
}

func isNodeReady(node *corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}
func nodeEndpointIP(node *corev1.Node) net.IP {
	var internal net.IP
	for _, a := range node.Status.Addresses {
		switch a.Type {
		case corev1.NodeExternalIP:
			return net.ParseIP(a.Address) // prefer external
		case corev1.NodeInternalIP:
			internal = net.ParseIP(a.Address)
		}
	}
	return internal
}

func GetPeers(client *kubernetes.Clientset) (peers []wg.Peer, err error) {
	nodes_all, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, node := range nodes_all.Items {
		if !isNodeReady(&node) || strings.ToLower(node.Name) == strings.ToLower(os.Getenv("NODE_NAME")) {
			continue
		}
		endpoint := nodeEndpointIP(&node)
		SecretName := fmt.Sprintf("%s-%s", SecretNameBase, node.Name)
		secret, err := client.CoreV1().Secrets(NS).Get(context.Background(), SecretName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		peerKey := secret.Data["pubKey"]
		pubKey, err := wg.ImportKey(peerKey)
		if err != nil {
			return nil, err
		}

		port, err := strconv.Atoi(os.Getenv("LISTEN_PORT"))
		if err != nil {
			return nil, err
		}
		peer := wg.Peer{
			Endpoint:   &net.UDPAddr{IP: endpoint, Port: port},
			PublicKey:  pubKey,
			AllowedIPs: []string{os.Getenv("CIDR")},
		}
		peers = append(peers, peer)
	}

	return peers, nil
}
