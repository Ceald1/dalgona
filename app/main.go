package main

//go:generate mkdir -p bpf
import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"os/signal"
	"syscall"
	"time"

	"sync"

	"context"

	log "charm.land/log/v2"
	network_bpf "github.com/Ceald1/dalgona/app/bpf/network"
	dK8s "github.com/Ceald1/dalgona/app/k8s_d"
	wireguard "github.com/Ceald1/dalgona/app/wireguard"
	"github.com/cilium/ebpf"
	K "k8s.io/client-go/kubernetes"
)

// reread keys from secrets
func ReloadKeys(client *K.Clientset, wgClient wireguard.Conf) wireguard.Conf {
	time.Sleep(1 * time.Second)
	keys, err := dK8s.GetPeers(client)
	if err != nil {
		log.Warn(err.Error())
		return wgClient
	}
	wgClient.Peers = keys

	return wgClient

}

// update the bpf map initially and then watch for changes
func GetPodIPs(client *K.Clientset, bpfHandle *network_bpf.NetworkObjects) {
	ips, err := dK8s.GetIPs(client)
	if err != nil {
		log.Warn(err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 1. Declare a named mutex so both the watcher and your loop share the exact same lock
	var mu sync.RWMutex

	// 2. Run the blocking foreground watcher in a background goroutine
	go func() {
		if err := dK8s.WatchIPs(client, &ips, &mu); err != nil {
			log.Warn(err)
			stop() // Cancel ctx if the watcher crashes so the foreground loop exits too
		}
	}()

	// 3. Run your "while" loop in the foreground until Ctrl+C / SIGTERM (or watcher error)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping loop and shutting down watcher...")
			return

		case <-ticker.C:
			// Safely read from the shared ips map while WatchIPs updates it
			mu.RLock()
			keys := make([]uint32, 0, len(ips))
			values := make([]network_bpf.NetworkStrKey, 0, len(ips))

			for k, v := range ips { // key is service, value is ip
				ip, err := ipv4ToBe32(v)
				if err != nil {
					log.Warn(err.Error())
					continue
				}

				var value network_bpf.NetworkStrKey
				maxCopy := len(k)
				if maxCopy >= len(value.Name) {
					maxCopy = len(value.Name) - 1 // keep at least one trailing 0 if C reads a NUL-terminated string
					log.Warnf("service name %q truncated to %d bytes", k, maxCopy)
				}
				for i := 0; i < maxCopy; i++ {
					value.Name[i] = int8(k[i])
				}

				keys = append(keys, ip)
				values = append(values, value)
			}

			if len(keys) > 0 {
				_, err := bpfHandle.IpsMap.BatchUpdate(keys, values, &ebpf.BatchOptions{
					Flags: uint64(ebpf.UpdateAny),
				})
				if err != nil {
					log.Warn(err)
				}
			}
			// Example: iterate over ips or copy them for processing
			//for k, v := range ips { // key is service, value is ip
			//	var value network_bpf.NetworkStrKey
			//	for i := 0; i < len(k); i++ {
			//		value.Name[i] = int8(k[i])
			//	}
			//	ip, err := ipv4ToBe32(v)
			//	if err != nil {
			//		log.Warn(err.Error())
			//		continue
			//	}

			//	err = bpfHandle.IpsMap.Update(ip, value, ebpf.UpdateAny)
			//	if err != nil {
			//		log.Warn(err)
			//	}
			//	//log.Infof("Current IP entry: %s -> %v", k, v)
			//}
			mu.RUnlock()
		}
	}
}

func ipv4ToBe32(ipStr string) (uint32, error) {
	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		return 0, fmt.Errorf("parse ip %q: %w", ipStr, err)
	}
	if !addr.Is4() {
		return 0, fmt.Errorf("not an IPv4 address: %q", ipStr)
	}

	ip4 := addr.As4() // [byte0, byte1, byte2, byte3] in network/big-endian order
	return binary.BigEndian.Uint32(ip4[:]), nil
}

//func GetPodIPs(client *K.Clientset) {
//	ips, err := dK8s.GetIPs(client)
//	if err != nil {
//		log.Warn(err.Error())
//	}
//	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
//	defer stop()
//	err = dK8s.WatchIPs(client, &ips, &sync.RWMutex{}, ctx)
//	if err != nil {
//		log.Warn(err)
//	}
//}
