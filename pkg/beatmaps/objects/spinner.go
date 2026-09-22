package objects

import (
	"github.com/Lekuruu/gosu/pkg/beatmaps/timing"
	"strconv"
)

type Spinner struct {
	*HitObject

	Timings *timing.Timings
}

func NewSpinner(data []string) *Spinner {
	spinner := &Spinner{
		HitObject: commonParse(data, 6),
	}

	spinner.EndTime, _ = strconv.ParseFloat(data[5], 64)

	return spinner
}

func (spinner *Spinner) SetTiming(timings *timing.Timings) {
	spinner.Timings = timings
}

func (spinner *Spinner) GetType() Type {
	return SPINNER
}
