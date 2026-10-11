package events

type Type uint8

const (
	TypeBackground Type = iota
	TypeVideo
	TypeBreak
	TypeColourTransformation
	TypeSprite
	TypeSample
	TypeAnimation

	TypeUnknown Type = 255
)

type Layer uint8

const (
	LayerBackground Layer = iota
	LayerFail
	LayerPass
	LayerForeground
	LayerOverlay
)

type Origin string

const (
	OriginTopLeft      Origin = "TopLeft"
	OriginTopCentre    Origin = "TopCentre"
	OriginTopRight     Origin = "TopRight"
	OriginCentreLeft   Origin = "CentreLeft"
	OriginCentre       Origin = "Centre"
	OriginCentreRight  Origin = "CentreRight"
	OriginBottomLeft   Origin = "BottomLeft"
	OriginBottomCentre Origin = "BottomCentre"
	OriginBottomRight  Origin = "BottomRight"
)

type AnimationLoop string

const (
	AnimationLoopForever AnimationLoop = "LoopForever"
	AnimationLoopOnce    AnimationLoop = "LoopOnce"
)

type Event interface {
	Type() Type
}

type Background struct {
	StartTime int64
	Filename  string
	X         float64
	Y         float64
}

func (*Background) Type() Type { return TypeBackground }

type Video struct {
	StartTime int64
	Filename  string
	X         float64
	Y         float64
}

func (*Video) Type() Type { return TypeVideo }

type Break struct {
	StartTime int64
	EndTime   int64
}

func (*Break) Type() Type { return TypeBreak }

func (event *Break) Duration() int64 {
	return event.EndTime - event.StartTime
}

type ColourTransformation struct {
	StartTime int64
	Red       int
	Green     int
	Blue      int
}

func (*ColourTransformation) Type() Type { return TypeColourTransformation }

// StoryboardObject contains the fields shared by sprites and animations.
type StoryboardObject struct {
	Layer    Layer
	Origin   Origin
	Filename string
	X        float64
	Y        float64
	Commands []*Command
}

type Sprite struct {
	StoryboardObject
}

func (*Sprite) Type() Type { return TypeSprite }

type Animation struct {
	StoryboardObject
	FrameCount int
	FrameDelay int64
	Loop       AnimationLoop
}

func (*Animation) Type() Type { return TypeAnimation }

type Sample struct {
	StartTime int64
	Layer     Layer
	Filename  string
	Volume    int
}

func (*Sample) Type() Type { return TypeSample }

// RawEvent preserves an event type that we don't know how to deal with.
type RawEvent struct {
	Raw string
}

func (*RawEvent) Type() Type { return TypeUnknown }

type Collection []Event

func (collection Collection) Backgrounds() []*Background {
	return collection.Filter[*Background]()
}

func (collection Collection) Videos() []*Video {
	return collection.Filter[*Video]()
}

func (collection Collection) Breaks() []*Break {
	return collection.Filter[*Break]()
}

func (collection Collection) Filter[T Event]() []T {
	var result []T

	for _, event := range collection {
		if value, ok := event.(T); ok {
			result = append(result, value)
		}
	}

	return result
}
