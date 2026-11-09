// procwatch is a small command-line tool that reports Linux system health.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/22wh1a1263/procwatch/internal/sysinfo"
)

const usage = `procwatch - Linux system health from /proc

Usage:
  procwatch snapshot [--json] [--disk PATH]
  procwatch top      [--json] [-n COUNT]
  procwatch ports    [--json]
  procwatch check    [--max-mem PCT] [--max-disk PCT] [--max-load N]
  procwatch serve    [--addr :8080]   web page and JSON API

Exit codes: 0 ok, 1 error, 2 check failed.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
	var err error
	code := 0
	switch os.Args[1] {
	case "snapshot":
		err = runSnapshot(os.Args[2:])
	case "top":
		err = runTop(os.Args[2:])
	case "ports":
		err = runPorts(os.Args[2:])
	case "check":
		code, err = runCheck(os.Args[2:])
	case "serve":
		err = runServe(os.Args[2:])
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func runSnapshot(args []string) error {
	fs := flag.NewFlagSet("snapshot", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "print JSON")
	disk := fs.String("disk", "/", "path whose filesystem to report")
	fs.Parse(args)
	s, err := sysinfo.CollectSnapshot("/proc", *disk)
	if err != nil {
		return err
	}
	if *asJSON {
		return printJSON(s)
	}
	fmt.Printf("uptime   %.1f hours\n", s.UptimeSeconds/3600)
	fmt.Printf("load     %.2f %.2f %.2f\n", s.Load.One, s.Load.Five, s.Load.Fifteen)
	fmt.Printf("memory   %.1f%% used (%d of %d MB)\n", s.Memory.UsedPercent, s.Memory.UsedKB/1024, s.Memory.TotalKB/1024)
	fmt.Printf("disk %s  %.1f%% used (%d GB free)\n", s.Disk.Path, s.Disk.UsedPercent, s.Disk.FreeBytes>>30)
	return nil
}

func runTop(args []string) error {
	fs := flag.NewFlagSet("top", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "print JSON")
	n := fs.Int("n", 10, "number of processes")
	fs.Parse(args)
	procs, err := sysinfo.TopProcesses("/proc", *n)
	if err != nil {
		return err
	}
	if *asJSON {
		return printJSON(procs)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "PID\tSTATE\tRSS(MB)\tNAME")
	for _, p := range procs {
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\n", p.PID, p.State, p.RSSKB/1024, p.Name)
	}
	return w.Flush()
}

func runPorts(args []string) error {
	fs := flag.NewFlagSet("ports", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "print JSON")
	fs.Parse(args)
	ports, err := sysinfo.ListeningPorts("/proc")
	if err != nil {
		return err
	}
	if *asJSON {
		return printJSON(ports)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "PORT\tADDRESS\tFAMILY")
	for _, p := range ports {
		fmt.Fprintf(w, "%d\t%s\t%s\n", p.Port, p.Address, p.Family)
	}
	return w.Flush()
}

func runCheck(args []string) (int, error) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	maxMem := fs.Float64("max-mem", 90, "fail if memory used percent exceeds this")
	maxDisk := fs.Float64("max-disk", 90, "fail if disk used percent exceeds this")
	maxLoad := fs.Float64("max-load", 0, "fail if 1-minute load exceeds this (0 = ignore)")
	fs.Parse(args)
	s, err := sysinfo.CollectSnapshot("/proc", "/")
	if err != nil {
		return 1, err
	}
	failed := false
	if s.Memory.UsedPercent > *maxMem {
		fmt.Printf("FAIL memory %.1f%% > %.1f%%\n", s.Memory.UsedPercent, *maxMem)
		failed = true
	}
	if s.Disk.UsedPercent > *maxDisk {
		fmt.Printf("FAIL disk %.1f%% > %.1f%%\n", s.Disk.UsedPercent, *maxDisk)
		failed = true
	}
	if *maxLoad > 0 && s.Load.One > *maxLoad {
		fmt.Printf("FAIL load %.2f > %.2f\n", s.Load.One, *maxLoad)
		failed = true
	}
	if failed {
		return 2, nil
	}
	fmt.Println("OK")
	return 0, nil
}
