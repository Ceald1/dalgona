package main

//go:generate mkdir -p bpf
import (
	log "charm.land/log/v2"
	dK8s "github.com/Ceald1/dalgona/app/k8s"
	network_bpf "github.com/Ceald1/dalgona/app/network"
)
