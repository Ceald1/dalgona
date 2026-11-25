package wireguard

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Conf struct {
	Peers      []Peer
	PrivateKey wgtypes.Key
	Port       int
}

type Peer struct {
	AllowedIPs []string
	PublicKey  wgtypes.Key
	Preshared  *wgtypes.Key
	Endpoint   *net.UDPAddr //required!!
}

const ifaceName = "dalgona_wg0"

// usage:
// priv, _ := wgtypes.GeneratePrivateKey()
// peerPriv, _ := wgtypes.GeneratePrivateKey() // normally the peer sends you its public key
//
//	c, err := wg.NewClient(wg.Conf{
//		PrivateKey: priv,
//		ListenPort: 51820,
//		Peers: []wg.Peer{
//			{PublicKey: peerPriv.PublicKey(), AllowedIPs: []string{"10.0.0.2/32"}},
//		},
//	})
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
// defer c.Close()
func NewClient(conf Conf) (c *wgctrl.Client, err error) {
	c, err = wgctrl.New()
	if err != nil {
		return
	}
	localAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	wg_cfg := wgtypes.Config{ReplacePeers: true, ListenPort: &conf.Port, PrivateKey: &conf.PrivateKey}
	for _, peer := range conf.Peers {
		var allowed []net.IPNet
		var peerConf wgtypes.PeerConfig
		for _, cidr := range peer.AllowedIPs {
			_, ipnet, err := net.ParseCIDR(cidr)
			if err != nil {
				c.Close()
				return nil, err
			}
			allowed = append(allowed, *ipnet)
		}
		found := false
		for _, addr := range localAddrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				found = ipnet.IP.Equal(peer.Endpoint.IP)
				if found != false {
					goto foundPeerAddress
				}
			}
		}

		peerConf = wgtypes.PeerConfig{PresharedKey: peer.Preshared, AllowedIPs: allowed, Endpoint: peer.Endpoint, ReplaceAllowedIPs: true, PublicKey: peer.PublicKey}
		wg_cfg.Peers = append(wg_cfg.Peers, peerConf)
	foundPeerAddress:
		continue
	}
	err = ensureInterface()
	if err != nil {
		return
	}
	err = c.ConfigureDevice(ifaceName, wg_cfg)
	return
}

func ensureInterface() error {
	if _, err := netlink.LinkByName(ifaceName); err == nil {
		return nil // already exists
	}
	link := &netlink.GenericLink{
		LinkAttrs: netlink.LinkAttrs{Name: ifaceName},
		LinkType:  "wireguard",
	}
	if err := netlink.LinkAdd(link); err != nil {
		return fmt.Errorf("create %s: %w", ifaceName, err)
	}
	return nil
}

func ImportKey(key []byte) (importedKey wgtypes.Key, err error) {
	importedKey, err = wgtypes.ParseKey(string(key))

	return
}
