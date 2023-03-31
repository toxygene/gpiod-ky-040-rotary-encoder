package device

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
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
	logger   *logrus.Entry
}

func NewRotaryEncoder(chip *gpiod.Chip, clockPin int, dataPin int, logger *logrus.Entry) *RotaryEncoder {
	return &RotaryEncoder{
		chip:     chip,
		clockPin: clockPin,
		dataPin:  dataPin,
		logger:   logger,
	}
}

func (t *RotaryEncoder) Run(ctx context.Context, actions chan<- Action) error {
	t.logger.Info("rotary encoder started")
	defer t.logger.Info("rotary encoder finished")

	previousClock, err := t.readClock()
	if err != nil {
		return fmt.Errorf("read clock: %w", err)
	}

	t.logger.WithField("clockPin", t.clockPin).WithField("value", previousClock).Trace("read clock")

	var lines *gpiod.Lines
	lineValues := []int{0, 0}

	handler := func(event gpiod.LineEvent) {
		t.logger.Info("rotary encoder event handler started")
		defer t.logger.Info("rotary encoder event handler finished")

		err := lines.Values(lineValues)
		if err != nil {
			t.logger.WithError(err).Error("read rotary encoder line values failed")
			panic(err)
		}

		t.logger.WithField("clockPin", t.clockPin).WithField("dataPin", t.dataPin).WithField("lineValues", lineValues).Trace("read rotary encoder line values")

		if previousClock != lineValues[0] && lineValues[0] == 1 {
			if lineValues[1] != lineValues[0] {
				actions <- CounterClockwise
			} else {
				actions <- Clockwise
			}
		}

		previousClock = lineValues[0]
	}

	lines, err = t.chip.RequestLines([]int{t.clockPin, t.dataPin}, gpiod.AsInput, gpiod.WithBothEdges, gpiod.WithEventHandler(handler))
	if err != nil {
		t.logger.WithError(err).WithField("clockPin", t.clockPin).WithField("dataPin", t.dataPin).Error("request rotary encoder lines failed")
		return fmt.Errorf("request rotary encoder lines: %w", err)
	}

	defer lines.Close()

	<-ctx.Done()

	return nil
}

func (t *RotaryEncoder) readClock() (int, error) {
	logger := t.logger.WithField("clockPin", t.clockPin)

	logger.Info("reading rotary encoder clock value")
	defer logger.Info("read rotary encoder clock value")

	clockLine, err := t.chip.RequestLine(t.clockPin, gpiod.AsInput)
	if err != nil {
		lineInfo, _ := t.chip.LineInfo(t.clockPin)

		logger.WithError(err).WithField("clockLineInfo", lineInfo).Error("request rotary encoder clock line failed")
		return 0, fmt.Errorf("request rotary encoder clock line: %w", err)
	}

	defer clockLine.Close()

	value, err := clockLine.Value()
	if err != nil {
		logger.WithError(err).WithField("clockLine", clockLine).Error("read rotary encoder clock value failed")
		return 0, fmt.Errorf("read rotary encoder clock value: %w", err)
	}

	logger.WithField("value", value).Trace("read rotary encoder clock value")

	return value, nil
}
