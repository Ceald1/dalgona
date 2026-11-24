package bpf

//go:generate mkdir -p ./network
//go:generate go tool bpf2go -cc clang -output-dir ./network -go-package network Network ../../bpf/network/network.c
