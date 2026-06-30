package alert_test

import (
	"testing"
	"time"

	"github.com/stdx-space/currency-bot/internal/alert"
	"github.com/stdx-space/currency-bot/internal/wise"
)

// makeRates builds a slice of hourly rates spread across distinct days.
// values[i] is the single rate for day i (UTC midnight + 1h).
func makeRates(values []float64) []wise.Rate {
	base := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
	rates := make([]wise.Rate, len(values))
	for i, v := range values {
		rates[i] = wise.Rate{
			Source: "HKD",
			Target: "CAD",
			Value:  v,
			Time:   base.AddDate(0, 0, i),
		}
	}
	return rates
}

func TestEvaluate_ModeAlways(t *testing.T) {
	rates := makeRates([]float64{1.0, 2.0, 3.0, 4.0, 5.0})
	result := alert.Evaluate(rates, alert.ModeAlways, 2)
	if !result.ShouldAlert {
		t.Error("ModeAlways: expected ShouldAlert=true")
	}
}

func TestEvaluate_Thresholds(t *testing.T) {
	// 5 days, one rate per day. Top 2 days = days with 4.0 and 5.0 → threshold hi = 4.0.
	// Bottom 2 days = days with 1.0 and 2.0 → threshold lo = 2.0.
	rates := makeRates([]float64{1.0, 2.0, 3.0, 4.0, 5.0})
	result := alert.Evaluate(rates, alert.ModeAlways, 2)

	if result.ThresholdHi != 4.0 {
		t.Errorf("ThresholdHi = %v, want 4.0", result.ThresholdHi)
	}
	if result.ThresholdLo != 2.0 {
		t.Errorf("ThresholdLo = %v, want 2.0", result.ThresholdLo)
	}
}

func TestEvaluate_MaximaDetection(t *testing.T) {
	tests := []struct {
		name        string
		values      []float64 // last value is "current"
		alertDays   int
		wantMaxima  bool
		wantMinima  bool
		wantAlert   bool
	}{
		{
			name:       "current is the highest",
			values:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			alertDays:  2,
			wantMaxima: true,
			wantMinima: false,
			wantAlert:  true,
		},
		{
			name:       "current is in the middle",
			values:     []float64{1.0, 2.0, 5.0, 4.0, 3.0},
			alertDays:  2,
			wantMaxima: false,
			wantMinima: false,
			wantAlert:  false,
		},
		{
			name:       "current is the lowest",
			values:     []float64{5.0, 4.0, 3.0, 2.0, 1.0},
			alertDays:  2,
			wantMaxima: false,
			wantMinima: true,
			wantAlert:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rates := makeRates(tt.values)
			result := alert.Evaluate(rates, alert.ModeMaxima, tt.alertDays)
			if result.IsMaxima != tt.wantMaxima {
				t.Errorf("IsMaxima = %v, want %v", result.IsMaxima, tt.wantMaxima)
			}
			if result.IsMinima != tt.wantMinima {
				t.Errorf("IsMinima = %v, want %v", result.IsMinima, tt.wantMinima)
			}
			if result.ShouldAlert != tt.wantAlert {
				t.Errorf("ShouldAlert = %v, want %v", result.ShouldAlert, tt.wantAlert)
			}
		})
	}
}

func TestEvaluate_ModeMinima(t *testing.T) {
	// current (1.0) is the lowest → should alert in minima mode
	rates := makeRates([]float64{5.0, 4.0, 3.0, 2.0, 1.0})
	result := alert.Evaluate(rates, alert.ModeMinima, 2)
	if !result.ShouldAlert {
		t.Error("ModeMinima: expected ShouldAlert=true when current is lowest")
	}
}

func TestEvaluate_AlertDaysClamped(t *testing.T) {
	// alertDays > number of days — should not panic, clamp to available days.
	rates := makeRates([]float64{1.0, 2.0, 3.0})
	result := alert.Evaluate(rates, alert.ModeMaxima, 10)
	if !result.ShouldAlert {
		t.Error("clamped alertDays: current should always be within all days window")
	}
}

func TestEvaluate_EmptyRates(t *testing.T) {
	result := alert.Evaluate(nil, alert.ModeAlways, 3)
	if result.ShouldAlert {
		t.Error("empty rates: expected ShouldAlert=false")
	}
}

func TestEvaluate_MultipleSameDayRates(t *testing.T) {
	// Multiple hourly entries on the same day — daily max should be used.
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rates := []wise.Rate{
		{Value: 1.0, Time: base.Add(1 * time.Hour)}, // day 1, low
		{Value: 3.0, Time: base.Add(2 * time.Hour)}, // day 1, high → daily max = 3.0
		{Value: 2.0, Time: base.AddDate(0, 0, 1)},   // day 2 → daily max = 2.0
		{Value: 4.0, Time: base.AddDate(0, 0, 2)},   // day 3, current
	}
	result := alert.Evaluate(rates, alert.ModeMaxima, 1)
	// Top 1 day = day 3 (4.0). Current = 4.0 → IsMaxima.
	if !result.IsMaxima {
		t.Errorf("IsMaxima = false, want true (current=4.0 >= thresholdHi=4.0)")
	}
}
