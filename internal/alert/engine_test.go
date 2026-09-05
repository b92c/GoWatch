package alert

import (
	"testing"
	"time"

	"github.com/b92c/gowatch/pkg/metrics"
)

func TestAlertEngineConsecutiveHits(t *testing.T) {
	engine := NewAlertEngine(nil)

	host := metrics.HostInfo{
		CPUCount: 1,
		MemTotal: 1024 * 1024 * 1024, // 1GB
	}

	c := ContainerEvaluatorData{
		ID:         "container-1",
		Service:    "web",
		State:      "running",
		CPUPercent: 90.0, // High CPU (threshold is 85%)
		MemUsage:   100 * 1024 * 1024,
	}

	// Amostra 1: Hit 1 -> Não deve disparar ainda (ConsecutiveHits = 3)
	alerts := engine.Evaluate([]ContainerEvaluatorData{c}, host)
	if len(alerts) != 0 {
		t.Fatalf("expected 0 active alerts on hit 1, got %d", len(alerts))
	}

	// Amostra 2: Hit 2 -> Não deve disparar ainda
	alerts = engine.Evaluate([]ContainerEvaluatorData{c}, host)
	if len(alerts) != 0 {
		t.Fatalf("expected 0 active alerts on hit 2, got %d", len(alerts))
	}

	// Amostra 3: Hit 3 -> DEVE disparar (StatusFIRING)
	alerts = engine.Evaluate([]ContainerEvaluatorData{c}, host)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 active alert on hit 3, got %d", len(alerts))
	}
	if alerts[0].Status != StatusFiring {
		t.Fatalf("expected status FIRING, got %s", alerts[0].Status)
	}
	if alerts[0].RuleID != "HIGH_CPU" {
		t.Fatalf("expected rule HIGH_CPU, got %s", alerts[0].RuleID)
	}
}

func TestAlertEngineHysteresis(t *testing.T) {
	engine := NewAlertEngine(nil)
	host := metrics.HostInfo{CPUCount: 1, MemTotal: 1000}

	c := ContainerEvaluatorData{
		ID:         "c1",
		Service:    "db",
		State:      "running",
		CPUPercent: 90.0,
	}

	// Disparar o alerta com 3 hits
	for i := 0; i < 3; i++ {
		engine.Evaluate([]ContainerEvaluatorData{c}, host)
	}

	alerts := engine.GetActiveAlertsSnapshot()
	if len(alerts) != 1 || alerts[0].Status != StatusFiring {
		t.Fatalf("alert should be FIRING")
	}

	// Queda para 80% (abaixo do Threshold 85%, mas acima da Histerese 75%)
	c.CPUPercent = 80.0
	engine.Evaluate([]ContainerEvaluatorData{c}, host)

	alerts = engine.GetActiveAlertsSnapshot()
	if len(alerts) != 1 || alerts[0].Status != StatusFiring {
		t.Fatalf("alert should STILL be FIRING due to hysteresis (val=80%%, hyst=75%%)")
	}

	// Queda para 70% (abaixo da Histerese 75%) -> DEVE resolver
	c.CPUPercent = 70.0
	engine.Evaluate([]ContainerEvaluatorData{c}, host)

	alerts = engine.GetActiveAlertsSnapshot()
	if len(alerts) != 1 || alerts[0].Status != StatusResolved {
		t.Fatalf("alert should be RESOLVED after dropping below hysteresis threshold")
	}
}

func TestAlertEngineSafeMemoryLimit(t *testing.T) {
	engine := NewAlertEngine(nil)

	// Host memTotal = 0 (simula erro de coleta ou limite não definido sem crash/div zero)
	host := metrics.HostInfo{CPUCount: 1, MemTotal: 0}
	c := ContainerEvaluatorData{
		ID:       "c-zero-mem",
		MemUsage: 5000,
	}

	alerts := engine.Evaluate([]ContainerEvaluatorData{c}, host)
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts for zero mem total host fallback, got %d", len(alerts))
	}
}

func TestAlertEngineStaleContainerCleanup(t *testing.T) {
	engine := NewAlertEngine(nil)
	host := metrics.HostInfo{CPUCount: 1, MemTotal: 1000}

	c := ContainerEvaluatorData{
		ID:        "c-ephemeral",
		OOMEvents: 1,
	}

	// 1 hit para OOM_KILLED (ConsecutiveHits = 1)
	engine.Evaluate([]ContainerEvaluatorData{c}, host)

	alerts := engine.GetActiveAlertsSnapshot()
	if len(alerts) != 1 || alerts[0].Status != StatusFiring {
		t.Fatalf("expected OOM alert firing")
	}

	// Na próxima avaliação, o container sumiu da lista
	engine.Evaluate([]ContainerEvaluatorData{}, host)

	alerts = engine.GetActiveAlertsSnapshot()
	if len(alerts) != 1 || alerts[0].Status != StatusResolved {
		t.Fatalf("expected stale container alert to be auto-resolved")
	}
}

func TestAlertEngineConcurrency(t *testing.T) {
	engine := NewAlertEngine(nil)
	host := metrics.HostInfo{CPUCount: 2, MemTotal: 8000}

	done := make(chan bool)
	go func() {
		for i := 0; i < 50; i++ {
			engine.Evaluate([]ContainerEvaluatorData{
				{ID: "c1", CPUPercent: 95.0},
			}, host)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			_ = engine.GetActiveAlertsSnapshot()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	<-done
	<-done
}
