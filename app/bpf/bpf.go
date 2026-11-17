package bpf

//go:generate mkdir ./network
//go:generate go tool bpf2go -cc clang -output-dir ./network -go-package network network ../../bpf/network/network.c
