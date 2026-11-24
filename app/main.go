package main

//go:generate mkdir -p bpf
import (
	"context"
	"encoding/binary"
	"fmt"
	rand "math/rand/v2"
	"net/netip"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	log "charm.land/log/v2"
	network_bpf "github.com/Ceald1/dalgona/app/bpf/network"
	dK8s "github.com/Ceald1/dalgona/app/k8s_d"
	wireguard "github.com/Ceald1/dalgona/app/wireguard"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/rlimit"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	K "k8s.io/client-go/kubernetes"
)

// ReloadKeys re-reads peers from the k8s secret into conf (mutex-protected).
func ReloadKeys(client *K.Clientset, conf *wireguard.Conf, mu *sync.Mutex, wg_cfg *wgctrl.Client) {
	peers, err := dK8s.GetPeers(client)
	if err != nil {
		log.Warn(err.Error())
		return
	}
	mu.Lock()
	conf.Peers = peers
	mu.Unlock()
	cfg, err := wireguard.NewClient(*conf)
	if err != nil {
		log.Fatal(err)
	}
	wg_cfg = cfg
}

// GetPodIPs syncs the pod IP map into BPF until ctx is cancelled.
func GetPodIPs(ctx context.Context, client *K.Clientset, bpfHandle *network_bpf.NetworkMaps) {
	ips, err := dK8s.GetIPs(client)
	if err != nil {
		log.Warn(err.Error())
	}
	if ips == nil {
		ips = make(map[string]string) // adjust type to match what GetIPs returns
	}

	var mu sync.RWMutex

	ctx, stop := context.WithCancel(ctx)
	defer stop()

	go func() {
		if err := dK8s.WatchIPs(client, &ips, &mu); err != nil {
			log.Warn(err)
			stop()
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	prev := map[uint32]struct{}{}

	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping loop and shutting down watcher...")
			return

		case <-ticker.C:
			mu.RLock()
			keys := make([]uint32, 0, len(ips))
			values := make([]network_bpf.NetworkStrKey, 0, len(ips))
			current := make(map[uint32]struct{}, len(ips))

			for name, ipStr := range ips { // key is service, value is ip
				ip, err := ipv4ToBe32(ipStr)
				if err != nil {
					log.Warn(err.Error())
					continue
				}

				var value network_bpf.NetworkStrKey
				maxCopy := len(name)
				if maxCopy >= len(value.Name) {
					maxCopy = len(value.Name) - 1
					log.Warnf("service name %q truncated to %d bytes", name, maxCopy)
				}
				for i := 0; i < maxCopy; i++ {
					value.Name[i] = int8(name[i])
				}

				keys = append(keys, ip)
				values = append(values, value)
				current[ip] = struct{}{}
			}
			mu.RUnlock()

			if len(keys) > 0 {
				if _, err := bpfHandle.IpsMap.BatchUpdate(keys, values, &ebpf.BatchOptions{
					Flags: uint64(ebpf.UpdateAny),
				}); err != nil {
					log.Warn(err)
				}
			}

			// remove entries that no longer exist
			for k := range prev {
				if _, ok := current[k]; !ok {
					if err := bpfHandle.IpsMap.Delete(k); err != nil {
						log.Warn(err)
					}
				}
			}
			prev = current
		}
	}
}

// ipv4ToBe32 returns a uint32 whose in-memory bytes are in network order,
// which is what a __be32 key in the BPF map expects.
func ipv4ToBe32(ipStr string) (uint32, error) {
	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		return 0, fmt.Errorf("parse ip %q: %w", ipStr, err)
	}
	if !addr.Is4() {
		return 0, fmt.Errorf("not an IPv4 address: %q", ipStr)
	}
	ip4 := addr.As4()
	return binary.NativeEndian.Uint32(ip4[:]), nil
}

func startPeriodicTask(ctx context.Context, interval time.Duration, client *K.Clientset, conf *wireguard.Conf, mu *sync.Mutex, cfg *wgctrl.Client) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ReloadKeys(client, conf, mu, cfg)
			}
		}
	}()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var conf wireguard.Conf
	privkey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		log.Fatal(err)
	}
	k8sClient, err := dK8s.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	if err := dK8s.UpdateSecret(k8sClient, nil, &privkey); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Duration(rand.IntN(30)) * time.Second)

	port, err := strconv.Atoi(os.Getenv("WG_PORT"))
	if err != nil {
		port = 51820 // default; err is intentionally not fatal
	}
	conf.Port = port
	conf.PrivateKey = privkey

	conf.Peers, err = dK8s.GetPeers(k8sClient)
	if err != nil {
		log.Fatal(err)
	}
	wgClient, err := wireguard.NewClient(conf)
	if err != nil {
		log.Fatal(err)
	}
	defer wgClient.Close()

	var confMu sync.Mutex
	startPeriodicTask(ctx, 2*time.Second, k8sClient, &conf, &confMu, wgClient)

	// load the BPF objects so IpsMap actually exists
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}
	var objs network_bpf.NetworkObjects
	if err := network_bpf.LoadNetworkObjects(&objs, nil); err != nil {
		log.Fatal(err)
	}
	defer objs.Close()

	GetPodIPs(ctx, k8sClient, &objs.NetworkMaps)
}
