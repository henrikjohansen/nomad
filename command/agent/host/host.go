package host

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type HostData struct {
	OS         string
	Network    string
	ResolvConf string
	Hosts      string
	Systemd    string
	Disk       map[string]DiskUsage
}

type DiskUsage struct {
	DiskMB int64
	UsedMB int64
}

func MakeHostData() (*HostData, error) {
	uname, err := uname()
	if err != nil {
		return nil, fmt.Errorf("error uname: %s", err.Error())
	}
	du := make(map[string]DiskUsage)
	for _, path := range mountedPaths() {
		u, err := diskUsage(path)
		if err != nil {
			continue
		}
		du[path] = u
	}

	return &HostData{
		OS:         uname,
		Network:    network(),
		ResolvConf: resolvConf(),
		Hosts:      etcHosts(),
		Systemd:    slurp("/etc/systemd/nomad.conf"),
		Disk:       du,
	}, nil
}

// diskUsage calculates the DiskUsage
func diskUsage(path string) (du DiskUsage, err error) {
	s, err := makeDf(path)
	if err != nil {
		return du, err
	}

	disk := float64(s.total())
	// Bavail is blocks available to unprivileged users, Bfree includes reserved blocks
	free := float64(s.available())
	used := disk - free
	mb := float64(1048576)

	disk = disk / mb
	used = used / mb

	du.DiskMB = int64(disk)
	du.UsedMB = int64(used)
	return du, nil
}

// call executes the command line and returns stdout, ignoring errors
func call(cmd string, args ...string) string {
	// cmd = which(cmd)
	out, _ := exec.Command(cmd, args...).Output()
	return string(out)
}

// slurp returns the file contents as a string, ignoring errors
func slurp(path string) string {
	var sb strings.Builder
	buf := make([]byte, 512)
	fh, err := os.Open(path)
	var l int
	for {
		l, err = fh.Read(buf)
		if err != nil {
			if l > 0 {
				sb.Write(buf[0 : l-1])
			}
			break
		}
		sb.Write(buf[0 : l-1])
	}
	return sb.String()
}
