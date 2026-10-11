// Package beatmaps parses osu! beatmap files and exposes their metadata,
// events, timing points, difficulty settings, and hit objects.
package beatmaps

import (
	"strconv"
	"strings"

	"github.com/Lekuruu/gosu/pkg/beatmaps/difficulty"
	"github.com/Lekuruu/gosu/pkg/beatmaps/events"
	"github.com/Lekuruu/gosu/pkg/beatmaps/objects"
	"github.com/Lekuruu/gosu/pkg/beatmaps/timing"
)

type Beatmap struct {
	FileVersion int
	Mode        int

	Artist        string
	ArtistUnicode string

	Title        string
	TitleUnicode string

	Version string
	Creator string
	Source  string
	Tags    string

	SliderMultiplier float64
	StackLeniency    float64

	Difficulty *difficulty.Difficulty

	Audio string

	SetID int64
	MapID int64

	PreviewTime int64

	Length   int
	Circles  int
	Sliders  int
	Spinners int

	BPMDurations map[float64]int64

	Timings    *timing.Timings
	Events     events.Collection
	HitObjects []objects.IHitObject

	arSpecified bool
}

func NewBeatmap() *Beatmap {
	return &Beatmap{
		Timings:       timing.NewTimings(),
		StackLeniency: 0.7,
		Difficulty:    difficulty.NewDifficulty(5, 5, 5, 5),
		BPMDurations:  make(map[float64]int64),
	}
}

func (beatmap *Beatmap) ParsePoint(point string) {
	line := strings.Split(point, ",")
	pointTime, _ := strconv.ParseFloat(line[0], 64)
	bpm, _ := strconv.ParseFloat(line[1], 64)

	signature := 4
	sampleSet := beatmap.Timings.BaseSet
	sampleIndex := 1
	sampleVolume := 1.0
	inherited := false
	kiai := false
	omitFirstBarLine := false

	if len(line) > 2 {
		signature, _ = strconv.Atoi(line[2])
		if signature == 0 {
			signature = 4
		}
	}

	if len(line) > 3 {
		sampleSet, _ = strconv.Atoi(line[3])
	}

	if len(line) > 4 {
		sampleIndex, _ = strconv.Atoi(line[4])
	}

	if len(line) > 5 {
		sV, _ := strconv.Atoi(line[5])
		sampleVolume = float64(sV) / 100
	}

	if len(line) > 6 {
		inh, _ := strconv.Atoi(line[6])
		inherited = inh == 0
	}

	if len(line) > 7 {
		ki, _ := strconv.Atoi(line[7])
		kiai = (ki & 1) > 0
		omitFirstBarLine = (ki & 8) > 0
	}

	beatmap.Timings.AddPoint(pointTime, bpm, sampleSet, sampleIndex, sampleVolume, signature, inherited, kiai, omitFirstBarLine)
}

func (beatmap *Beatmap) FinalizePoints() {
	beatmap.Timings.FinalizePoints()
	beatmap.BPMDurations = beatmap.Timings.CalculateBPMDurations(int64(beatmap.Length))
}

// CommonBPM returns the BPM which is active for the longest duration.
func (beatmap *Beatmap) CommonBPM() float64 {
	longestDuration := int64(-1)
	commonBPM := 0.0

	for bpm, duration := range beatmap.BPMDurations {
		isHigher := duration > longestDuration
		isSameButLowerBPM := duration == longestDuration && bpm < commonBPM

		if isHigher || isSameButLowerBPM {
			commonBPM = bpm
			longestDuration = duration
		}
	}
	return commonBPM
}

// MinBPM returns the lowest BPM throughout the beatmap.
func (beatmap *Beatmap) MinBPM() float64 {
	minimum := 0.0
	for bpm := range beatmap.BPMDurations {
		if minimum == 0 || bpm < minimum {
			minimum = bpm
		}
	}
	return minimum
}

// MaxBPM returns the highest BPM throughout the beatmap.
func (beatmap *Beatmap) MaxBPM() float64 {
	maximum := 0.0
	for bpm := range beatmap.BPMDurations {
		maximum = max(maximum, bpm)
	}
	return maximum
}

// TotalLength returns the beatmap's total length in seconds.
func (beatmap *Beatmap) TotalLength() float64 {
	if len(beatmap.HitObjects) <= 1 {
		return 0
	}

	lastObject := beatmap.HitObjects[len(beatmap.HitObjects)-1]
	return max(lastObject.GetStartTime(), 0) / 1000
}

// DrainLength returns the beatmap's drain length in seconds, i.e.
// the time from the first hit object to the last, excluding breaks.
func (beatmap *Beatmap) DrainLength() float64 {
	if len(beatmap.HitObjects) <= 1 {
		return 0
	}
	firstObject := beatmap.HitObjects[0]
	lastObject := beatmap.HitObjects[len(beatmap.HitObjects)-1]

	breakTime := 0.0
	for _, breakEvent := range beatmap.Events.Breaks() {
		breakTime += float64(breakEvent.Duration())
	}

	drainLength := lastObject.GetStartTime() - firstObject.GetStartTime() - breakTime
	return max(drainLength, 0) / 1000
}

// HasStoryboardElements reports whether the beatmap contains a sprite or animation event.
func (beatmap *Beatmap) HasStoryboardElements() bool {
	for _, event := range beatmap.Events {
		if event == nil {
			continue
		}

		eventType := event.Type()
		if eventType == events.TypeSprite {
			return true
		}
		if eventType == events.TypeAnimation {
			return true
		}
	}
	return false
}
