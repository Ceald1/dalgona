vagrant up --provider=libvirt
sudo virsh net-update vagrant-libvirt add ip-dhcp-host '<host mac="52:54:00:aa:bb:10" ip="192.168.121.10"/>' --live --config
sudo virsh net-update vagrant-libvirt add ip-dhcp-host '<host mac="52:54:00:aa:bb:11" ip="192.168.121.11"/>' --live --config
sudo virsh net-update vagrant-libvirt add ip-dhcp-host '<host mac="52:54:00:aa:bb:12" ip="192.168.121.12"/>' --live --config

vagrant ssh k3s-server -c "sudo systemctl reboot"
vagrant ssh k3s-agent1 -c "sudo systemctl reboot"
vagrant ssh k3s-agent2 -c "sudo systemctl reboot"

sleep 10


vagrant ssh k3s-server -c "sudo apt update -y ;sudo apt install build-essential git curl -y"
vagrant ssh k3s-agent1 -c "sudo apt update -y ;sudo apt install build-essential git curl -y"
vagrant ssh k3s-agent2 -c "sudo apt update -y ;sudo apt install build-essential git curl -y"


k3sup install \
  --ip 192.168.121.10 \
  --user vagrant \
  --ssh-key .vagrant/machines/k3s-server/libvirt/private_key \
  --cluster \
  --tls-san 192.168.121.10 \
  --k3s-extra-args '--node-ip=192.168.121.10 --flannel-iface=eth0' \
  --local-path ~/.kube/config-dalgona \
  --context dalgona


SERVER_KEY=.vagrant/machines/k3s-server/libvirt/private_key
PUB=$(ssh-keygen -y -f "$SERVER_KEY")

for n in 1 2; do
  vagrant ssh k3s-agent$n -c "echo '$PUB' >> ~/.ssh/authorized_keys"
done

for n in 1 2; do
  k3sup join \
    --server \
    --ip 192.168.121.1$n \
    --server-ip 192.168.121.10 \
    --user vagrant \
    --server-user vagrant \
    --ssh-key "$SERVER_KEY" \
    --k3s-extra-args "--node-ip=192.168.121.1$n --flannel-iface=eth0 --tls-san 192.168.121.10"
done

#for n in 1 2; do
#  k3sup join \
#    --ip 192.168.121.1$n \
#    --server-ip 192.168.121.10 \
#    --user vagrant \
#    --ssh-key .vagrant/machines/k3s-agent$n/libvirt/private_key \
#    --k3s-extra-args "--node-ip=192.168.121.1$n --flannel-iface=eth0"
#done
