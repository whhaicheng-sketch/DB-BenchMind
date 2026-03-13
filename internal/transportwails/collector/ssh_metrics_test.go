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
