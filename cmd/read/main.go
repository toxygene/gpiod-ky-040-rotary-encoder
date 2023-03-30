package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/toxygene/gpiod-ky-040-rotary-encoder/device"
	"github.com/warthog618/gpiod"
	"golang.org/x/sync/errgroup"
)

func main() {
	chipName := flag.String("chipName", "", "Chip name for the GPIO device of the rotary encoder and button")
	help := flag.Bool("help", false, "print help page")
	logging := flag.String("logging", "", "logging level")
	rotaryEncoderClockPinNumber := flag.Int("rotaryEncoderClock", 0, "GPIO number of clock pin for the rotary encoder")
	rotaryEncoderDataPinNumber := flag.Int("rotaryEncoderData", 0, "GPIO number of data pin for the rotary encoder")

	flag.Parse()

	if *help || *rotaryEncoderClockPinNumber == 0 || *rotaryEncoderDataPinNumber == 0 {
		flag.Usage()

		if *help {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	logger := logrus.New()

	if *logging != "" {
		logLevel, err := logrus.ParseLevel(*logging)
		if err != nil {
			println(fmt.Errorf("parse log level: %w", err).Error())
			os.Exit(1)
		}

		logger.SetLevel(logLevel)
	}

	actions := make(chan device.Action)

	chip, err := gpiod.NewChip(*chipName)
	if err != nil {
		logger.WithError(err).WithField("chipName", *chipName).Error("created gpio chip failed")
		panic(fmt.Errorf("create gpiod.chip: %w", err))
	}

	defer chip.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	defer close(c)

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		logger.Info("starting interrupt handler")
		defer logger.Info("interrupt handler finished")

		select {
		case <-c:
			return errors.New("application interrupted")
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	g.Go(func() error {
		defer close(actions)

		logger.Info("starting rotary encoder goroutine")
		defer logger.Info("rotary encoder goroutine finished")

		re := device.NewRotaryEncoder(chip, *rotaryEncoderClockPinNumber, *rotaryEncoderDataPinNumber, logrus.NewEntry(logger))

		return re.Run(ctx, actions)
	})

	g.Go(func() error {
		logger.Info("starting actions handler")
		defer logger.Info("actions handler finished")

		i := 0
		for action := range actions {
			logger.WithField("action", action).Trace("received action")

			switch action {
			case device.Clockwise:
				i++
			case device.CounterClockwise:
				i--
			}

			println(i)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		panic(fmt.Errorf("running application goroutines: %w", err))
	}
}
