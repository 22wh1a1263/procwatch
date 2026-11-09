package sysinfo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
)

// Disk is filesystem usage for one path.
type Disk struct {
	Path        string  `json:"path"`
	TotalBytes  uint64  `json:"total_bytes"`
	FreeBytes   uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

// Snapshot is a point-in-time view of the machine.
type Snapshot struct {
	UptimeSeconds float64 `json:"uptime_seconds"`
	Load          Load    `json:"load"`
	Memory        Memory  `json:"memory"`
	Disk          Disk    `json:"disk"`
}

func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

// CollectSnapshot reads /proc and the filesystem at diskPath.
func CollectSnapshot(procRoot, diskPath string) (Snapshot, error) {
	var s Snapshot
	t, err := readFile(filepath.Join(procRoot, "uptime"))
	if err != nil {
		return s, err
	}
	if s.UptimeSeconds, err = ParseUptime(t); err != nil {
		return s, err
	}
	if t, err = readFile(filepath.Join(procRoot, "loadavg")); err != nil {
		return s, err
	}
	if s.Load, err = ParseLoadavg(t); err != nil {
		return s, err
	}
	if t, err = readFile(filepath.Join(procRoot, "meminfo")); err != nil {
		return s, err
	}
	if s.Memory, err = ParseMeminfo(t); err != nil {
		return s, err
	}
	if s.Disk, err = DiskUsage(diskPath); err != nil {
		return s, err
	}
	return s, nil
}

// DiskUsage returns usage for the filesystem containing path.
func DiskUsage(path string) (Disk, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return Disk{}, fmt.Errorf("statfs %s: %w", path, err)
	}
	total := st.Blocks * uint64(st.Bsize)
	free := st.Bavail * uint64(st.Bsize)
	d := Disk{Path: path, TotalBytes: total, FreeBytes: free}
	if total > 0 {
		d.UsedPercent = float64(total-free) * 100 / float64(total)
	}
	return d, nil
}

// TopProcesses returns the n processes using the most resident memory.
// Processes that exit while being read are skipped.
func TopProcesses(procRoot string, n int) ([]Process, error) {
	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return nil, err
	}
	procs := []Process{}
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		t, err := readFile(filepath.Join(procRoot, e.Name(), "status"))
		if err != nil {
			continue
		}
		p, err := ParseStatus(pid, t)
		if err != nil {
			continue
		}
		procs = append(procs, p)
	}
	sort.Slice(procs, func(i, j int) bool {
		if procs[i].RSSKB != procs[j].RSSKB {
			return procs[i].RSSKB > procs[j].RSSKB
		}
		return procs[i].PID < procs[j].PID
	})
	if n > 0 && len(procs) > n {
		procs = procs[:n]
	}
	return procs, nil
}

// ListeningPorts returns TCP sockets in LISTEN state (IPv4 and IPv6), sorted by port.
func ListeningPorts(procRoot string) ([]Port, error) {
	all := []Port{}
	for _, fam := range []string{"tcp", "tcp6"} {
		t, err := readFile(filepath.Join(procRoot, "net", fam))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		ps, err := ParseNetTCP(t, fam)
		if err != nil {
			return nil, err
		}
		all = append(all, ps...)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Port < all[j].Port })
	return all, nil
}
