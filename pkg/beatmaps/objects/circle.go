package objects

import (
	"github.com/Lekuruu/gosu/pkg/beatmaps/difficulty"
)

type Circle struct {
	*HitObject

	diff *difficulty.Difficulty
}

func NewCircle(data []string) *Circle {
	return &Circle{
		HitObject: commonParse(data, 5),
	}
}

func (circle *Circle) GetType() Type {
	return CIRCLE
}
