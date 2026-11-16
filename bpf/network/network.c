#include "network.h"

static __always_inline bool is_ip_in_cidr(__be32 ip, __be32 net, __be32 mask) {
  return (ip & mask) == (net & mask);
}

SEC("tc") int to_wg(struct __sk_buff *skb) {
  void *data = (void *)(long)skb->data;
  void *end = (void *)(long)skb->data_end;
  struct ethhdr *eth = data;
  struct iphdr *ip;

  if ((void *)(eth + 1) > end || eth->h_proto != bpf_htons(ETH_P_IP))
    return TC_ACT_OK;
  ip = (void *)(eth + 1);
  if ((void *)(ip + 1) > end)
    return TC_ACT_OK;

  /* skip wg's own UDP so it isn't encrypted twice (ctx_is_wireguard) */
  if (ip->protocol == IPPROTO_UDP) {
    if (ip->ihl < 5)
      return TC_ACT_OK;
    struct udphdr *udp = (void *)ip + ip->ihl * 4;
    if ((void *)(udp + 1) > end)
      return TC_ACT_OK;
    if (udp->dest == bpf_htons(WG_PORT) && udp->source == udp->dest)
      return TC_ACT_OK;
  }

  __be32 *net_be = bpf_map_lookup_elem(&net_info, &NET);
  __be32 *mask_be = bpf_map_lookup_elem(&net_info, &MASK);
  if (!net_be || !mask_be) /* <-- this is what's missing */
    return TC_ACT_OK;

  if (!is_ip_in_cidr(ip->daddr, *net_be, *mask_be)) {
    return TC_ACT_OK;
  }

  __u32 *wg_ifindex = bpf_map_lookup_elem(&wg_info, &WGI);
  if (!wg_ifindex) {
    return TC_ACT_OK;
  }
  return bpf_redirect(*wg_ifindex, 0);
}
