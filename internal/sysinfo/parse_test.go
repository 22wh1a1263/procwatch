package sysinfo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMeminfo(t *testing.T) {
	text := "MemTotal:       16000000 kB\nMemFree:         1000000 kB\nMemAvailable:    4000000 kB\n"
	m, err := ParseMeminfo(text)
	if err != nil {
		t.Fatal(err)
	}
	if m.UsedKB != 12000000 || m.UsedPercent != 75 {
		t.Fatalf("got %+v", m)
	}
}

func TestParseMeminfoMissingField(t *testing.T) {
	if _, err := ParseMeminfo("MemTotal: 100 kB\n"); err == nil {
		t.Fatal("expected error for missing MemAvailable")
	}
}

func TestParseMeminfoAvailableTooLarge(t *testing.T) {
	if _, err := ParseMeminfo("MemTotal: 100 kB\nMemAvailable: 200 kB\n"); err == nil {
		t.Fatal("expected error when available > total")
	}
}

func TestParseLoadavg(t *testing.T) {
	l, err := ParseLoadavg("0.52 0.58 0.59 1/467 12345\n")
	if err != nil || l.One != 0.52 || l.Fifteen != 0.59 {
		t.Fatalf("got %+v, %v", l, err)
	}
	if _, err := ParseLoadavg("0.5"); err == nil {
		t.Fatal("expected error for short input")
	}
	if _, err := ParseLoadavg("a b c"); err == nil {
		t.Fatal("expected error for non-numeric input")
	}
}

func TestParseUptime(t *testing.T) {
	v, err := ParseUptime("350735.47 234388.90\n")
	if err != nil || v != 350735.47 {
		t.Fatalf("got %v, %v", v, err)
	}
}

func TestParseStatus(t *testing.T) {
	text := "Name:\tbash\nState:\tS (sleeping)\nVmRSS:\t    3456 kB\n"
	p, err := ParseStatus(42, text)
	if err != nil || p.Name != "bash" || p.State != "S" || p.RSSKB != 3456 {
		t.Fatalf("got %+v, %v", p, err)
	}
}

func TestParseStatusKernelThreadHasNoRSS(t *testing.T) {
	p, err := ParseStatus(2, "Name:\tkthreadd\nState:\tS (sleeping)\n")
	if err != nil || p.RSSKB != 0 {
		t.Fatalf("got %+v, %v", p, err)
	}
}

func TestParseNetTCP(t *testing.T) {
	text := `  sl  local_address rem_address   st tx_queue rx_queue
   0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000
   1: 0100007F:0050 0100007F:C350 01 00000000:00000000 00:00000000 00000000
   2: 00000000:0016 00000000:0000 0A 00000000:00000000 00:00000000 00000000
`
	ps, err := ParseNetTCP(text, "tcp")
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 2 {
		t.Fatalf("want 2 listening sockets, got %d: %+v", len(ps), ps)
	}
	if ps[0].Address != "127.0.0.1" || ps[0].Port != 8080 {
		t.Fatalf("got %+v", ps[0])
	}
	if ps[1].Address != "0.0.0.0" || ps[1].Port != 22 {
		t.Fatalf("got %+v", ps[1])
	}
}

func TestParseNetTCP6(t *testing.T) {
	text := "  sl  local_address rem_address st\n   0: 00000000000000000000000000000000:01BB 00000000000000000000000000000000:0000 0A x\n"
	ps, err := ParseNetTCP(text, "tcp6")
	if err != nil || len(ps) != 1 || ps[0].Port != 443 || ps[0].Address != "::" {
		t.Fatalf("got %+v, %v", ps, err)
	}
}

func TestTopProcessesSortsAndSkipsBadEntries(t *testing.T) {
	root := t.TempDir()
	write := func(pid, body string) {
		dir := filepath.Join(root, pid)
		os.MkdirAll(dir, 0o755)
		os.WriteFile(filepath.Join(dir, "status"), []byte(body), 0o644)
	}
	write("10", "Name:\ta\nState:\tS\nVmRSS:\t100 kB\n")
	write("11", "Name:\tb\nState:\tR\nVmRSS:\t900 kB\n")
	write("12", "garbage without name\n")
	os.MkdirAll(filepath.Join(root, "self"), 0o755) // non-numeric entry is ignored
	ps, err := TopProcesses(root, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 2 || ps[0].Name != "b" {
		t.Fatalf("got %+v", ps)
	}
	one, _ := TopProcesses(root, 1)
	if len(one) != 1 {
		t.Fatalf("limit not applied: %+v", one)
	}
}

func TestDiskUsage(t *testing.T) {
	d, err := DiskUsage("/")
	if err != nil || d.TotalBytes == 0 || d.UsedPercent < 0 || d.UsedPercent > 100 {
		t.Fatalf("got %+v, %v", d, err)
	}
	if _, err := DiskUsage("/definitely/not/a/path"); err == nil {
		t.Fatal("expected error for missing path")
	}
}
