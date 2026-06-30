package alert

import (
	"sort"
	"time"

	"github.com/stommydx/narwhl/currency-bot/internal/wise"
)

type Mode string

const (
	ModeAlways Mode = "always"
	ModeMaxima Mode = "maxima"
	ModeMinima Mode = "minima"
)

type Result struct {
	Current     float64
	DailyMax    map[string]float64 // date string -> max rate for that day
	ThresholdHi float64            // Xth-highest daily rate (maxima threshold)
	ThresholdLo float64            // Xth-lowest daily rate (minima threshold)
	IsMaxima    bool               // current >= ThresholdHi
	IsMinima    bool               // current <= ThresholdLo
	ShouldAlert bool
}

// Evaluate analyses rates and decides whether an alert should be fired.
// alertDays: the X in "within highest/lowest X days".
func Evaluate(rates []wise.Rate, mode Mode, alertDays int) Result {
	if len(rates) == 0 {
		return Result{}
	}

	// Compute per-day max rate (use UTC date as key).
	dailyMax := make(map[string]float64)
	for _, r := range rates {
		day := r.Time.UTC().Format(time.DateOnly)
		if r.Value > dailyMax[day] {
			dailyMax[day] = r.Value
		}
	}

	// Sort daily maxima descending (for hi) and ascending (for lo).
	values := make([]float64, 0, len(dailyMax))
	for _, v := range dailyMax {
		values = append(values, v)
	}
	sort.Float64s(values) // ascending

	// Clamp alertDays to the number of days available.
	if alertDays > len(values) {
		alertDays = len(values)
	}

	thresholdHi := values[len(values)-alertDays] // Xth-highest (ascending: index from end)
	thresholdLo := values[alertDays-1]            // Xth-lowest

	// Current rate is the most recent entry.
	current := rates[len(rates)-1].Value

	isMaxima := current >= thresholdHi
	isMinima := current <= thresholdLo

	shouldAlert := false
	switch mode {
	case ModeAlways:
		shouldAlert = true
	case ModeMaxima:
		shouldAlert = isMaxima
	case ModeMinima:
		shouldAlert = isMinima
	}

	return Result{
		Current:     current,
		DailyMax:    dailyMax,
		ThresholdHi: thresholdHi,
		ThresholdLo: thresholdLo,
		IsMaxima:    isMaxima,
		IsMinima:    isMinima,
		ShouldAlert: shouldAlert,
	}
}
