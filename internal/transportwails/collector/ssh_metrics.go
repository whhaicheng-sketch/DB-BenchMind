package collector

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"golang.org/x/crypto/ssh"
)

type SSHMetricPoint struct {
	Timestamp    int64   `json:"timestamp"`
	CPUUser      float64 `json:"cpu_user"`
	CPUSys       float64 `json:"cpu_sys"`
	CPUIOWait    float64 `json:"cpu_iowait"`
	DiskReadBps  float64 `json:"disk_read_bps"`
	DiskWriteBps float64 `json:"disk_write_bps"`
}

type SSHMetricsCollector struct {
	config   *connection.SSHTunnelConfig
	interval time.Duration

	mu      sync.RWMutex
	running bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	history []SSHMetricPoint
	lastCPU *cpuCounters
	lastDisk *diskCounters
}

type cpuCounters struct {
	user   uint64
	nice   uint64
	system uint64
	idle   uint64
	iowait uint64
	irq    uint64
	softirq uint64
	steal  uint64
}

type diskCounters struct {
	readSectors  uint64
	writeSectors uint64
}

func NewSSHMetricsCollector(config *connection.SSHTunnelConfig, interval time.Duration) *SSHMetricsCollector {
	if interval <= 0 {
		interval = time.Second
	}
	return &SSHMetricsCollector{
		config:   config,
		interval: interval,
		history:  make([]SSHMetricPoint, 0, 300),
	}
}

func (c *SSHMetricsCollector) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.running {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.running = true
	c.wg.Add(1)
	go c.loop(ctx)
	return nil
}

func (c *SSHMetricsCollector) Stop() {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return
	}
	c.running = false
	cancel := c.cancel
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	c.wg.Wait()
}

func (c *SSHMetricsCollector) Snapshot() []SSHMetricPoint {
	c.mu.RLock()
	defer c.mu.RUnlock()
	points := make([]SSHMetricPoint, len(c.history))
	copy(points, c.history)
	return points
}

func (c *SSHMetricsCollector) loop(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	c.collect()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

func (c *SSHMetricsCollector) collect() {
	output, err := c.exec("cat /proc/stat; echo '---'; cat /proc/diskstats")
	if err != nil {
		return
	}
	cpuNext, diskNext, err := parseLinuxMetrics(output)
	if err != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().UnixMilli()
	point := SSHMetricPoint{Timestamp: now}
	if c.lastCPU != nil {
		point.CPUUser, point.CPUSys, point.CPUIOWait = calcCPUPercentages(*c.lastCPU, cpuNext)
	}
	if c.lastDisk != nil {
		seconds := c.interval.Seconds()
		if seconds <= 0 {
			seconds = 1
		}
		point.DiskReadBps = float64(delta(diskNext.readSectors, c.lastDisk.readSectors)*512) / seconds
		point.DiskWriteBps = float64(delta(diskNext.writeSectors, c.lastDisk.writeSectors)*512) / seconds
	}
	c.lastCPU = &cpuNext
	c.lastDisk = &diskNext
	c.history = append(c.history, point)
	if len(c.history) > 300 {
		c.history = append([]SSHMetricPoint(nil), c.history[len(c.history)-300:]...)
	}
}

func (c *SSHMetricsCollector) exec(cmd string) (string, error) {
	sshConfig, err := c.configToSSH()
	if err != nil {
		return "", err
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.config.Host, c.config.Port), sshConfig)
	if err != nil {
		return "", err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	out, err := session.Output(cmd)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (c *SSHMetricsCollector) configToSSH() (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User:            c.config.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	if c.config.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(c.config.Password))
	}
	if len(config.Auth) == 0 {
		return nil, fmt.Errorf("ssh password is required")
	}
	return config, nil
}

func parseLinuxMetrics(output string) (cpuCounters, diskCounters, error) {
	parts := strings.Split(output, "---")
	if len(parts) != 2 {
		return cpuCounters{}, diskCounters{}, fmt.Errorf("unexpected ssh metrics output")
	}
	cpu, err := parseCPUCounters(parts[0])
	if err != nil {
		return cpuCounters{}, diskCounters{}, err
	}
	disk, err := parseDiskCounters(parts[1])
	if err != nil {
		return cpuCounters{}, diskCounters{}, err
	}
	return cpu, disk, nil
}

func parseCPUCounters(raw string) (cpuCounters, error) {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 8 && fields[0] == "cpu" {
			values := make([]uint64, 8)
			for i := 0; i < 8; i++ {
				v, err := strconv.ParseUint(fields[i+1], 10, 64)
				if err != nil {
					return cpuCounters{}, err
				}
				values[i] = v
			}
			return cpuCounters{
				user: values[0], nice: values[1], system: values[2], idle: values[3],
				iowait: values[4], irq: values[5], softirq: values[6], steal: values[7],
			}, nil
		}
	}
	return cpuCounters{}, fmt.Errorf("cpu counters not found")
}

func parseDiskCounters(raw string) (diskCounters, error) {
	re := regexp.MustCompile(`^(sd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme\d+n\d+)$`)
	var counters diskCounters
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 14 || !re.MatchString(fields[2]) {
			continue
		}
		readSectors, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			return diskCounters{}, err
		}
		writeSectors, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			return diskCounters{}, err
		}
		counters.readSectors += readSectors
		counters.writeSectors += writeSectors
	}
	return counters, nil
}

func calcCPUPercentages(prev, next cpuCounters) (float64, float64, float64) {
	prevIdle := prev.idle + prev.iowait
	nextIdle := next.idle + next.iowait
	prevNonIdle := prev.user + prev.nice + prev.system + prev.irq + prev.softirq + prev.steal
	nextNonIdle := next.user + next.nice + next.system + next.irq + next.softirq + next.steal
	prevTotal := prevIdle + prevNonIdle
	nextTotal := nextIdle + nextNonIdle
	totald := float64(delta(nextTotal, prevTotal))
	if totald <= 0 {
		return 0, 0, 0
	}
	user := float64(delta(next.user+next.nice, prev.user+prev.nice)) / totald * 100
	sys := float64(delta(next.system+next.irq+next.softirq, prev.system+prev.irq+prev.softirq)) / totald * 100
	iowait := float64(delta(next.iowait, prev.iowait)) / totald * 100
	return user, sys, iowait
}

func delta(next, prev uint64) uint64 {
	if next < prev {
		return 0
	}
	return next - prev
}
