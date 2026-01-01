package main

import (
	"context"
	"errors"
	"testing"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
)

type MockStatsGatherer struct {
	DiskUsage *disk.UsageStat
	DiskError error
	LoadAvg   *load.AvgStat
	LoadError error
	NumCPU    int
}

func (m *MockStatsGatherer) GetDiskUsage(path string) (*disk.UsageStat, error) {
	return m.DiskUsage, m.DiskError
}

func (m *MockStatsGatherer) GetLoadAvg() (*load.AvgStat, error) {
	return m.LoadAvg, m.LoadError
}

func (m *MockStatsGatherer) GetNumCPU() int {
	return m.NumCPU
}

func TestAnalyze_HealthySystem(t *testing.T) {
	mock := &MockStatsGatherer{
		DiskUsage: &disk.UsageStat{UsedPercent: 50.0},
		LoadAvg:   &load.AvgStat{Load5: 1.0},
		NumCPU:    4,
	}

	p := &SystemProvider{Gatherer: mock}

	report, _ := p.Analyze(context.Background(), nil)

	if len(report.Checks) != 2 {
		t.Fatalf("Expected 2 checks, got %d", len(report.Checks))
	}

	if !report.Checks[0].Passed {
		t.Error("Disk check failed, expected pass (50% usage)")
	}
	if report.Checks[0].Score != 10 {
		t.Error("Disk score incorrect")
	}

	if !report.Checks[1].Passed {
		t.Error("Load check failed, expected pass (Low load)")
	}
}

func TestAnalyze_UnhealthySystem(t *testing.T) {
	mock := &MockStatsGatherer{
		DiskUsage: &disk.UsageStat{UsedPercent: 95.0},
		LoadAvg:   &load.AvgStat{Load5: 10.0},
		NumCPU:    2,
	}

	p := &SystemProvider{Gatherer: mock}

	report, _ := p.Analyze(context.Background(), nil)

	if report.Checks[0].Passed {
		t.Error("Disk check passed unexpectedly (95% usage)")
	}
	if report.Checks[0].Score != 0 {
		t.Errorf("Disk check gave score %d, expected 0", report.Checks[0].Score)
	}

	if report.Checks[1].Passed {
		t.Error("Load check passed unexpectedly (High load)")
	}
}

func TestAnalyze_HardwareErrors(t *testing.T) {
	mock := &MockStatsGatherer{
		DiskError: errors.New("disk failure"),
		LoadError: errors.New("load read error"),
	}

	p := &SystemProvider{Gatherer: mock}

	report, _ := p.Analyze(context.Background(), nil)

	if len(report.Checks) != 0 {
		t.Errorf("Expected 0 checks on hardware failure, got %d", len(report.Checks))
	}
}
