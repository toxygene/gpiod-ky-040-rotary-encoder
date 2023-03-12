package device

import (
	"context"
	"fmt"

	"github.com/warthog618/gpiod"
)

type Action string

var (
	Clockwise        Action = "clockwise"
	CounterClockwise Action = "counterclockwise"
	Click            Action = "click"
)

type RotaryEncoder struct {
	chip     *gpiod.Chip
	clockPin int
	dataPin  int
}

func NewRotaryEncoder(chip *gpiod.Chip, clockPin int, dataPin int) *RotaryEncoder {
	return &RotaryEncoder{
		chip:     chip,
		clockPin: clockPin,
		dataPin:  dataPin,
	}
}

func (r *RotaryEncoder) Run(ctx context.Context, actions chan<- Action) error {
	previousClock, err := r.readClock()
	if err != nil {
		return fmt.Errorf("read clock: %w", err)
	}

	var lines *gpiod.Lines
	lineValues := []int{0, 0}

	handler := func(event gpiod.LineEvent) {
		err := lines.Values(lineValues)
		if err != nil {
			panic(err)
		}

		if previousClock != lineValues[0] && lineValues[0] == 1 {
			if lineValues[1] != lineValues[0] {
				actions <- Clockwise
			} else {
				actions <- CounterClockwise
			}
		}

		previousClock = lineValues[0]
	}

	lines, err = r.chip.RequestLines([]int{r.clockPin, r.dataPin}, gpiod.AsInput, gpiod.WithBothEdges, gpiod.WithEventHandler(handler))
	if err != nil {
		return fmt.Errorf("request lines: %w", err)
	}

	defer lines.Close()

	<-ctx.Done()

	return nil
}

func (r *RotaryEncoder) readClock() (int, error) {
	clockLine, err := r.chip.RequestLine(r.clockPin, gpiod.AsInput)
	if err != nil {
		return 0, fmt.Errorf("request clock line: %w", err)
	}

	defer clockLine.Close()

	value, err := clockLine.Value()
	if err != nil {
		return 0, fmt.Errorf("read clock value: %w", err)
	}

	return value, nil
}
