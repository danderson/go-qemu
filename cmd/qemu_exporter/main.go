package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/digitalocean/go-qemu/hypervisor"
	"github.com/digitalocean/go-qemu/qemuexporter"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	telemetryAddr = flag.String("telemetry.addr", ":9137", "host:port for QEMU exporter")
	telemetryPath = flag.String("telemetry.path", "/metrics", "URL path for surfacing collected metrics")

	qemuAddr = flag.String("qemu.addr", "qemu:///system", "QEMU monitor address")
)

func main() {
	flag.Parse()

	d := hypervisor.NewRPCDriver(func() (net.Conn, error) {
		return net.Dial("unix", "/var/run/libvirt/libvirt-sock")
	})

	hv := hypervisor.New(d)

	// Ensure the hypervisor is reachable
	doms, err := hv.DomainNames()
	if err != nil {
		log.Fatalf("failed to connect to QEMU hypervisor: %v", err)
	}

	log.Println(doms)

	// Register our exporter
	prometheus.MustRegister(qemuexporter.New(hv))

	// register http handlers
	http.Handle(*telemetryPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, *telemetryPath, http.StatusMovedPermanently)
	})

	log.Printf("starting QEMU exporter on %q", *telemetryAddr)

	if err := http.ListenAndServe(*telemetryAddr, nil); err != nil {
		log.Fatalf("unexpected failure of QEMU exporter HTTP server: %v", err)
	}
}
