#define BPF_NO_GLOBAL_DATA
#include "../vmlinux.h"
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

#define TC_ACT_OK 0
#define ETH_P_IP 0x0800
#define TC_ACT_SHOT 2
// enum {
//   IPPROTO_IP = 0,
//   IPPROTO_ICMP = 1,
//   IPPROTO_IGMP = 2,
//   IPPROTO_IPIP = 4,
//   IPPROTO_TCP = 6,
//   IPPROTO_EGP = 8,
//   IPPROTO_PUP = 12,
//   IPPROTO_UDP = 17,
//   IPPROTO_IDP = 22,
//   IPPROTO_TP = 29,
//   IPPROTO_DCCP = 33,
//   IPPROTO_IPV6 = 41,
//   IPPROTO_RSVP = 46,
//   IPPROTO_GRE = 47,
//   IPPROTO_ESP = 50,
//   IPPROTO_AH = 51,
//   IPPROTO_MTP = 92,
//   IPPROTO_BEETPH = 94,
//   IPPROTO_ENCAP = 98,
//   IPPROTO_PIM = 103,
//   IPPROTO_COMP = 108,
//   IPPROTO_L2TP = 115,
//   IPPROTO_SCTP = 132,
//   IPPROTO_UDPLITE = 136,
//   IPPROTO_MPLS = 137,
//   IPPROTO_ETHERNET = 143,
//   IPPROTO_AGGFRAG = 144,
//   IPPROTO_RAW = 255,
//   IPPROTO_SMC = 256,
//   IPPROTO_MPTCP = 262,
//   IPPROTO_MAX = 263,
// };

#if defined(bpf_strstr)
#else
extern int bpf_strstr(const char *str, const char *substr) __ksym;
#endif
#define ARGV_LEN 154
#define MAX_LEN 256
#define HOST_CIDR

typedef unsigned int u32;
typedef int pid_t;
const pid_t pid_filter = 0;

struct str_key {
  char name[MAX_LEN];
} str_key;

char LICENSE[] SEC("license") = "Dual BSD/GPL";
#define WG_PORT 51820
struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, __be32);
  __type(value, str_key);
  __uint(max_entries, 300000);
} ips_map SEC(".maps");

struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, str_key);
  __type(value, __be32);
  __uint(max_entries, 2);

} net_info SEC(".maps");

const char NET[MAX_LEN] = "net";
const char MASK[MAX_LEN] = "mask";

struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, str_key);
  __type(value, __u32);
  __uint(max_entries, 1);
} wg_info SEC(".maps");

const char WGI[MAX_LEN] = "wg_index";

// struct key_network {
//   char str[MAX_LEN];
// };
//
// struct value_network {
//   __be32 ip;
// };
//
// struct {
//   __uint(type, BPF_MAP_TYPE_HASH);
//   __uint(max_entries, 300000);
//   __type(key, struct key_network);
//   __type(value, struct value_network);
// } ips_map SEC(".maps");
