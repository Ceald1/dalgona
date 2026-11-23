package main

//go:generate mkdir -p bpf
import (
	"time"

	log "charm.land/log/v2"
	network_bpf "github.com/Ceald1/dalgona/app/bpf/network"
	dK8s "github.com/Ceald1/dalgona/app/k8s"
	wireguard "github.com/Ceald1/dalgona/app/wireguard"
	K "k8s.io/client-go/kubernetes"
)

// regenerate keys after 5 minutes
func ReloadKeys(client *K.Clientset, wgClient wireguard.Conf) {
	time.Sleep(5 * time.Minute)

}
