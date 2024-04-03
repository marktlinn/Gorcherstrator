package worker

import "github.com/c9s/goprocinfo/linux"

// Stats represents pointers to all the Linux processes information required to provide metrics about containers running on the system.
type Stats struct {
	MemStats  *linux.MemInfo
	CPUStats  *linux.CPUStat
	LoadStats *linux.LoadAvg
	DiskStats *linux.Disk
}

// Provides the total amount of memory in KB.
// Equivalent of MemTotal in /proc/meminfo.
func (s *Stats) MemTotalKB() uint64 {
	return s.MemStats.MemTotal
}

// Provides the total available memory for allocation.
// Equivalent of MemAvailable in /proc/meminfo.
func (s *Stats) MemAvailableKB() uint64 {
	return s.MemStats.MemAvailable
}

// Shows the total amount of memory used as a percentage of total memory.
func (s *Stats) MemUsedPercent() uint64 {
	return s.MemStats.MemAvailable / s.MemStats.MemTotal
}

// Shows the total amount of memory used in KB.
func (s *Stats) MemUsedKB() uint64 {
	return s.MemStats.MemTotal - s.MemStats.MemAvailable
}

// DiskFree returns the total amount of Disk space is free to be used.
func (s *Stats) DiskFree() uint64 {
	return s.DiskStats.Free
}

// DiskUsed returns the total amount of Disk space that's is being used.
func (s *Stats) DiskUsed() uint64 {
	return s.DiskStats.Used
}

// DiskTotal returns the total amount of Disk space that there is, including that which is currently used and that which is free.
func (s *Stats) DiskTotal() uint64 {
	return s.DiskStats.All
}

// CpuUsage gives the total amount of CPU currently being used as a percentage of the overall CPU capacity.
// The percentage is calucated as:
//
//	((Sum all states) - (sum of idle states)) / sum of all states
func (s *Stats) CpuUsage() float64 {
	idleStates := s.CPUStats.Idle + s.CPUStats.IOWait
	nonIdleStates := s.CPUStats.User + s.CPUStats.Nice + s.CPUStats.Steal + s.CPUStats.System + s.CPUStats.SoftIRQ + s.CPUStats.IRQ

	ttl := idleStates + nonIdleStates

	if ttl == 0 {
		return 0.00
	}

	return (float64(ttl) - float64(idleStates)/float64(ttl))
}
