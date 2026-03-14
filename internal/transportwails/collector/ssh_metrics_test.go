package collector

import "testing"

func TestParseLinuxMetrics(t *testing.T) {
	raw := "cpu  100 0 50 200 25 0 0 0 0 0\n---\n   8       0 sda 1 0 10 0 2 0 20 0 0 0 0 0 0 0 0 0 0\n"
	cpu, disk, err := parseLinuxMetrics(raw)
	if err != nil {
		t.Fatalf("parseLinuxMetrics() error = %v", err)
	}
	if cpu.user != 100 || cpu.system != 50 || cpu.iowait != 25 {
		t.Fatalf("unexpected cpu counters: %+v", cpu)
	}
	if disk.readSectors != 10 || disk.writeSectors != 20 {
		t.Fatalf("unexpected disk counters: %+v", disk)
	}
}

func TestCalcCPUPercentagesIncludesSteal(t *testing.T) {
	prev := cpuCounters{
		user: 100, system: 40, idle: 200, iowait: 20, irq: 0, softirq: 0, steal: 10,
	}
	next := cpuCounters{
		user: 130, system: 55, idle: 230, iowait: 30, irq: 0, softirq: 0, steal: 25,
	}

	user, sys, iowait, steal := calcCPUPercentages(prev, next)

	if user <= 0 {
		t.Fatalf("expected user percentage > 0, got %.2f", user)
	}
	if sys <= 0 {
		t.Fatalf("expected sys percentage > 0, got %.2f", sys)
	}
	if iowait <= 0 {
		t.Fatalf("expected iowait percentage > 0, got %.2f", iowait)
	}
	if steal <= 0 {
		t.Fatalf("expected steal percentage > 0, got %.2f", steal)
	}
}
