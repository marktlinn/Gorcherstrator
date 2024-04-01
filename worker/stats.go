package worker

import "github.com/c9s/goprocinfo/linux"

// Stats represents pointers to all the Linux processes information required to provide metrics about containers running on the system.
type Stats struct {
	MemStats  *linux.MemInfo
	CPUStats  *linux.CPUInfo
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
