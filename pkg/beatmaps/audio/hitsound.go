package audio

type HitSoundInfo struct {
	SampleSet    int
	AdditionSet  int
	CustomIndex  int
	CustomVolume float64
}

type HitSound int

const (
	Normal HitSound = 1 << iota
	Whistle
	Finish
	Clap
)
