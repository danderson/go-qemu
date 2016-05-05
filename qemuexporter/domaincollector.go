package qemuexporter

import (
	"log"
	"sync"

	"github.com/digitalocean/go-qemu"
	"github.com/digitalocean/go-qemu/hypervisor"

	"github.com/prometheus/client_golang/prometheus"
)

// A DomainCollector is a Prometheus collector for ZFS domains metrics.
type DomainCollector struct {
	hv *hypervisor.Hypervisor

	Domains *prometheus.Desc

	BlockReadBytes  *prometheus.Desc
	BlockWriteBytes *prometheus.Desc

	BlockReadOperations  *prometheus.Desc
	BlockWriteOperations *prometheus.Desc
	BlockFlushOperations *prometheus.Desc

	BlockReadTimeNanoseconds  *prometheus.Desc
	BlockWriteTimeNanoseconds *prometheus.Desc
	BlockFlushTimeNanoseconds *prometheus.Desc
	BlockIdleTimeNanoseconds  *prometheus.Desc
}

// NewDomainCollector creates a new DomainCollector.
func NewDomainCollector(hv *hypervisor.Hypervisor) *DomainCollector {
	const (
		subsystem = "domains"
	)

	labels := []string{
		"domain",
		"device",
	}

	return &DomainCollector{
		Domains: prometheus.NewDesc(
			// Subsystem is used as name so we get "qemu_domains"
			prometheus.BuildFQName(namespace, "", subsystem),
			"Total number of domains",
			nil,
			nil,
		),

		BlockReadBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "block_read_bytes_total"),
			"Number of bytes read from block device",
			labels,
			nil,
		),

		BlockWriteBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "block_write_bytes_total"),
			"Number of bytes written to block device",
			labels,
			nil,
		),

		BlockReadOperations: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "block_read_operations_total"),
			"Number of read operations from block device",
			labels,
			nil,
		),

		BlockWriteOperations: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "block_write_operations_total"),
			"Number of write operations to block device",
			labels,
			nil,
		),

		BlockFlushOperations: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "block_flush_operations_total"),
			"Number of flush operations to block device",
			labels,
			nil,
		),

		BlockReadTimeNanoseconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "block_read_time_nanoseconds_total"),
			"Time in nanoseconds spent reading from block device",
			labels,
			nil,
		),

		BlockWriteTimeNanoseconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "block_write_time_nanoseconds_total"),
			"Time in nanoseconds spent writing to block device",
			labels,
			nil,
		),

		BlockFlushTimeNanoseconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "block_flush_time_nanoseconds_total"),
			"Time in nanoseconds spent flushing to block device",
			labels,
			nil,
		),

		BlockIdleTimeNanoseconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "block_idle_time_nanoseconds_total"),
			"Time in nanoseconds the block device has spent idle",
			labels,
			nil,
		),

		hv: hv,
	}
}

// collect begins a metrics collection task for all metrics related to UniFi
// stations.
func (c *DomainCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	domains, err := c.hv.Domains()
	if err != nil {
		return c.Domains, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.Domains,
		prometheus.GaugeValue,
		float64(len(domains)),
	)

	var wg sync.WaitGroup
	wg.Add(len(domains))

	for _, d := range domains {
		go func(d *qemu.Domain) {
			defer func() {
				_ = d.Close()
				wg.Done()
			}()

			if err := c.collectBlockStats(ch, d); err != nil {
				log.Println(err)
				return
			}
		}(d)
	}

	wg.Wait()

	if err := c.hv.Disconnect(); err != nil {
		log.Printf("ERROR DISCONNECT: %#v", err)
		return c.BlockWriteBytes, err
	}

	return nil, nil
}

func (c *DomainCollector) collectBlockStats(ch chan<- prometheus.Metric, d *qemu.Domain) error {
	stats, err := d.BlockStats()
	if err != nil {
		return err
	}

	for _, s := range stats {
		labels := []string{
			d.Name,
			s.Device,
		}

		ch <- prometheus.MustNewConstMetric(
			c.BlockReadBytes,
			prometheus.CounterValue,
			float64(s.ReadBytes),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BlockWriteBytes,
			prometheus.CounterValue,
			float64(s.WriteBytes),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BlockReadOperations,
			prometheus.CounterValue,
			float64(s.ReadOperations),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BlockWriteOperations,
			prometheus.CounterValue,
			float64(s.WriteOperations),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BlockFlushOperations,
			prometheus.CounterValue,
			float64(s.FlushOperations),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BlockReadTimeNanoseconds,
			prometheus.CounterValue,
			float64(s.ReadTotalTimeNanoseconds),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BlockWriteTimeNanoseconds,
			prometheus.CounterValue,
			float64(s.WriteTotalTimeNanoseconds),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BlockFlushTimeNanoseconds,
			prometheus.CounterValue,
			float64(s.FlushTotalTimeNanoseconds),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BlockIdleTimeNanoseconds,
			prometheus.CounterValue,
			float64(s.IdleTimeNanoseconds),
			labels...,
		)
	}

	return nil
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *DomainCollector) Describe(ch chan<- *prometheus.Desc) {
	m := []*prometheus.Desc{
		c.Domains,
		c.BlockReadBytes,
		c.BlockWriteBytes,
		c.BlockReadOperations,
		c.BlockWriteOperations,
		c.BlockFlushOperations,
		c.BlockReadTimeNanoseconds,
		c.BlockWriteTimeNanoseconds,
		c.BlockFlushTimeNanoseconds,
		c.BlockIdleTimeNanoseconds,
	}

	for _, d := range m {
		ch <- d
	}
}

// Collect sends the metric values for each metric pertaining to ZFS domainss
// over to the provided prometheus Metric channel.
func (c *DomainCollector) Collect(ch chan<- prometheus.Metric) {
	if desc, err := c.collect(ch); err != nil {
		log.Printf("[ERROR] failed collecting domain metric %v: %v", desc, err)
		ch <- prometheus.NewInvalidMetric(desc, err)
		return
	}
}
