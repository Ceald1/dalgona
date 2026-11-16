# dalgona
kubernetes service mesh that uses wireguard and bpfs for secure cross node communication




## Notes
1. all services are obtained on each node
2. bpf maps are updated with services' IPs and or hostnames
3. when communication to an external service is detected it is routed through wireguard

### Updating services
1. using the informer client
```sh
go get k8s.io/client-go@latest k8s.io/api@latest k8s.io/apimachinery@latest
```
2. update existing maps or just expand them (worried about maps being too big)
3. repeat


