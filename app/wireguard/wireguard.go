package wireguard

import (
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
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
	err = c.ConfigureDevice("dalgona_wg0", wg_cfg)
	return
}

func ImportKey(key []byte) (importedKey wgtypes.Key, err error) {
	importedKey, err = wgtypes.ParseKey(string(key))

	return
}
