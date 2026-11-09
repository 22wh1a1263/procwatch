// Package sysinfo reads system information from the Linux /proc filesystem.
// Every parser takes plain text, so it can be tested without a real system.
package sysinfo

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// Memory holds values from /proc/meminfo, in kilobytes.
type Memory struct {
	TotalKB     uint64  `json:"total_kb"`
	AvailableKB uint64  `json:"available_kb"`
	UsedKB      uint64  `json:"used_kb"`
	UsedPercent float64 `json:"used_percent"`
}

// ParseMeminfo parses the text of /proc/meminfo.
func ParseMeminfo(text string) (Memory, error) {
	var m Memory
	var haveTotal, haveAvail bool
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		key, rest, ok := strings.Cut(sc.Text(), ":")
		if !ok {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		val, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		switch key {
		case "MemTotal":
			m.TotalKB, haveTotal = val, true
		case "MemAvailable":
			m.AvailableKB, haveAvail = val, true
		}
	}
	if !haveTotal || !haveAvail {
		return m, fmt.Errorf("meminfo: MemTotal or MemAvailable missing")
	}
	if m.AvailableKB > m.TotalKB {
		return m, fmt.Errorf("meminfo: available (%d) exceeds total (%d)", m.AvailableKB, m.TotalKB)
	}
	m.UsedKB = m.TotalKB - m.AvailableKB
	if m.TotalKB > 0 {
		m.UsedPercent = float64(m.UsedKB) * 100 / float64(m.TotalKB)
	}
	return m, nil
}

// Load holds the 1, 5 and 15 minute load averages.
type Load struct {
	One     float64 `json:"one"`
	Five    float64 `json:"five"`
	Fifteen float64 `json:"fifteen"`
}

// ParseLoadavg parses the text of /proc/loadavg.
func ParseLoadavg(text string) (Load, error) {
	f := strings.Fields(text)
	if len(f) < 3 {
		return Load{}, fmt.Errorf("loadavg: expected at least 3 fields, got %d", len(f))
	}
	var vals [3]float64
	for i := 0; i < 3; i++ {
		v, err := strconv.ParseFloat(f[i], 64)
		if err != nil {
			return Load{}, fmt.Errorf("loadavg: field %d: %w", i, err)
		}
		vals[i] = v
	}
	return Load{vals[0], vals[1], vals[2]}, nil
}

// ParseUptime parses /proc/uptime and returns seconds since boot.
func ParseUptime(text string) (float64, error) {
	f := strings.Fields(text)
	if len(f) < 1 {
		return 0, fmt.Errorf("uptime: empty input")
	}
	return strconv.ParseFloat(f[0], 64)
}

// Process is a summary of one process from /proc/<pid>/status.
type Process struct {
	PID   int    `json:"pid"`
	Name  string `json:"name"`
	State string `json:"state"`
	RSSKB uint64 `json:"rss_kb"`
}

// ParseStatus parses the text of /proc/<pid>/status.
func ParseStatus(pid int, text string) (Process, error) {
	p := Process{PID: pid}
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		key, rest, ok := strings.Cut(sc.Text(), ":")
		if !ok {
			continue
		}
		rest = strings.TrimSpace(rest)
		switch key {
		case "Name":
			p.Name = rest
		case "State":
			if rest != "" {
				p.State = string(rest[0])
			}
		case "VmRSS":
			if f := strings.Fields(rest); len(f) > 0 {
				p.RSSKB, _ = strconv.ParseUint(f[0], 10, 64)
			}
		}
	}
	if p.Name == "" {
		return p, fmt.Errorf("status for pid %d: no Name field", pid)
	}
	return p, nil
}

// Port is a listening TCP socket.
type Port struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Family  string `json:"family"`
}

const tcpListen = "0A"

// ParseNetTCP parses /proc/net/tcp or /proc/net/tcp6 and returns sockets in LISTEN state.
func ParseNetTCP(text, family string) ([]Port, error) {
	var out []Port
	sc := bufio.NewScanner(strings.NewReader(text))
	first := true
	for sc.Scan() {
		if first { // header line
			first = false
			continue
		}
		f := strings.Fields(sc.Text())
		if len(f) < 4 {
			continue
		}
		if f[3] != tcpListen {
			continue
		}
		addr, port, err := splitHexAddr(f[1], family)
		if err != nil {
			return nil, err
		}
		out = append(out, Port{Address: addr, Port: port, Family: family})
	}
	return out, nil
}

func splitHexAddr(s, family string) (string, int, error) {
	host, portHex, ok := strings.Cut(s, ":")
	if !ok {
		return "", 0, fmt.Errorf("bad address %q", s)
	}
	port, err := strconv.ParseUint(portHex, 16, 16)
	if err != nil {
		return "", 0, fmt.Errorf("bad port in %q: %w", s, err)
	}
	if family == "tcp" {
		if len(host) != 8 {
			return "", 0, fmt.Errorf("bad ipv4 in %q", s)
		}
		// /proc stores IPv4 as a little-endian hex word.
		var b [4]uint64
		for i := 0; i < 4; i++ {
			v, err := strconv.ParseUint(host[i*2:i*2+2], 16, 8)
			if err != nil {
				return "", 0, err
			}
			b[3-i] = v
		}
		return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3]), int(port), nil
	}
	if host == strings.Repeat("0", 32) {
		return "::", int(port), nil
	}
	return "ipv6", int(port), nil
}
