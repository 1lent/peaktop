package collector

import (
	"time"

	"github.com/brodie/peaktop/internal/apple"
	"github.com/brodie/peaktop/internal/types"
)

const diskCollectorName = "disk"

type DiskCollector struct {
	stats       types.DiskStats
	prevRead    uint64
	prevWrite   uint64
	prevTime    time.Time
	initialized bool
}

func NewDiskCollector() *DiskCollector {
	return &DiskCollector{}
}

func (c *DiskCollector) Name() string {
	return diskCollectorName
}

func (c *DiskCollector) Collect() error {
	readBytes, writeBytes, err := apple.GetDiskStats()
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

	if !c.initialized {
		c.prevRead = readBytes
		c.prevWrite = writeBytes
		c.prevTime = now
		c.initialized = true
		c.stats = types.DiskStats{}
		return nil
	}

	readDelta := safeUint64Delta(readBytes, c.prevRead)
	writeDelta := safeUint64Delta(writeBytes, c.prevWrite)

	readBps := float64(readDelta) / elapsed
	writeBps := float64(writeDelta) / elapsed

	iopRate := 0.0
	if elapsed > 0 {
		iopRate = (readBps + writeBps) / 4096
	}

	c.prevRead = readBytes
	c.prevWrite = writeBytes
	c.prevTime = now

	c.stats = types.DiskStats{
		ReadBytesPerSec:  readBps,
		WriteBytesPerSec: writeBps,
		IOPS:             iopRate,
	}

	return nil
}

func (c *DiskCollector) Stats() types.DiskStats {
	return c.stats
}
