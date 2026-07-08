package collector

import (
	"time"

	"github.com/brodie/peaktop/internal/apple"
	"github.com/brodie/peaktop/internal/types"
)

const networkCollectorName = "network"

type NetworkCollector struct {
	stats     types.NetworkStats
	prevCounters map[string]networkCounter
	prevTime     time.Time
	initialized  bool
}

type networkCounter struct {
	rxBytes uint64
	txBytes uint64
}

func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{
		prevCounters: make(map[string]networkCounter),
	}
}

func (c *NetworkCollector) Name() string {
	return networkCollectorName
}

func (c *NetworkCollector) Collect() error {
	interfaces, err := apple.GetNetworkStats()
	if err != nil {
		return err
	}

	now := time.Now()
	elapsed := 1.0
	if c.initialized {
		elapsed = now.Sub(c.prevTime).Seconds()
		if elapsed < 0.001 {
			elapsed = 0.001
		}
	}

	currentCounters := make(map[string]networkCounter, len(interfaces))
	var totalRxBps, totalTxBps float64
	var netInterfaces []types.NetInterface

	for _, iface := range interfaces {
		currentCounters[iface.Name] = networkCounter{
			rxBytes: iface.RxBytes,
			txBytes: iface.TxBytes,
		}

		if !c.initialized {
			netInterfaces = append(netInterfaces, types.NetInterface{
				Name:  iface.Name,
				RxBps: 0,
				TxBps: 0,
			})
			continue
		}

		prev, exists := c.prevCounters[iface.Name]
		rxDelta := safeUint64Delta(iface.RxBytes, prev.rxBytes)
		txDelta := safeUint64Delta(iface.TxBytes, prev.txBytes)

		rxBps := float64(rxDelta) / elapsed
		txBps := float64(txDelta) / elapsed

		totalRxBps += rxBps
		totalTxBps += txBps

		_ = exists
		netInterfaces = append(netInterfaces, types.NetInterface{
			Name:  iface.Name,
			RxBps: rxBps,
			TxBps: txBps,
		})
	}

	c.prevCounters = currentCounters
	c.prevTime = now
	c.initialized = true

	c.stats = types.NetworkStats{
		RxBytesPerSec: totalRxBps,
		TxBytesPerSec: totalTxBps,
		Interfaces:    netInterfaces,
	}

	return nil
}

func (c *NetworkCollector) Stats() types.NetworkStats {
	return c.stats
}

func safeUint64Delta(current, previous uint64) uint64 {
	if current >= previous {
		return current - previous
	}
	return 0
}
