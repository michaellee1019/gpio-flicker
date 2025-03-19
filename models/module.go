package models

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
	"go.viam.com/utils/rpc"
)

var (
	GpioFlicker      = resource.NewModel("michaellee1019", "gpio-flicker", "gpio-flicker")
	errUnimplemented = errors.New("unimplemented")
)

func init() {
	resource.RegisterService(generic.API, GpioFlicker,
		resource.Registration[resource.Resource, *Config]{
			Constructor: newGpioFlickerGpioFlicker,
		},
	)
}

type Config struct {
	Boards []BoardConfig `json:"boards"`
	Interval int `json:"interval_ms"`
}

type BoardConfig struct {
	Board string `json:"board"`
	Pins  []string `json:"pins"`
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns implicit dependencies based on the config.
// The path is the JSON path in your robot's config (not the `Config` struct) to the
// resource being validated; e.g. "components.0".
func (cfg *Config) Validate(path string) ([]string, error) {
	// Add config validation code here

	boardNames := make([]string, len(cfg.Boards))
	for i, board := range cfg.Boards {
		if board.Board == "" {
			return nil, fmt.Errorf("board is required on board number %d", i)
		}
		boardNames[i] = board.Board
	}

	return boardNames, nil
}

type gpioFlickerGpioFlicker struct {
	name resource.Name

	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()

	boards []*board.Board
	pins   []board.GPIOPin
	interval time.Duration

	// Uncomment this if the model does not have any goroutines that
	// need to be shut down while closing.
	// resource.TriviallyCloseable

}

func newGpioFlickerGpioFlicker(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (resource.Resource, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	res := &gpioFlickerGpioFlicker{
		name:       rawConf.ResourceName(),
		logger:     logger,
		cfg:        conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}

	if err := res.Reconfigure(ctx, deps, rawConf); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *gpioFlickerGpioFlicker) Name() resource.Name {
	return s.name
}

func (s *gpioFlickerGpioFlicker) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	s.logger.Error("reconfiguring with: ", s.cfg)

	brds := make([]*board.Board, len(s.cfg.Boards))
	var pins []board.GPIOPin

	for i, brdConf := range s.cfg.Boards {
		brd, err := board.FromDependencies(deps, brdConf.Board)
		if err != nil {
			return err
		}
		brds[i] = &brd

		for _, pinConf := range brdConf.Pins {
			pin, err := brd.GPIOPinByName(pinConf)
			if err != nil {
				return err
			}
			pins = append(pins, pin)
		}
	}

	s.boards = brds
	s.pins = pins

	s.interval = time.Duration(s.cfg.Interval) * time.Millisecond

	s.logger.Info("reconfigured")

	// Start the flicker routine
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.cancelCtx.Done():
				// Clean up when context is cancelled
				// Use background context for cleanup to ensure it completes
				cleanupCtx := context.Background()
				for _, pin := range s.pins {
					_ = pin.Set(cleanupCtx, false, nil)
				}
				return
			case <-ticker.C:
				// Randomly select and toggle one pin
				s.logger.Info("ticked")
				randomPin := s.pins[rand.Intn(len(s.pins))]

				currentValue, err := randomPin.Get(s.cancelCtx, nil)
				if err != nil {
					s.logger.Error("Failed to get pin state: ", "error", err)
				}
				// Use the cancelCtx instead of the parent ctx
				if err := randomPin.Set(s.cancelCtx, !currentValue, nil); err != nil {
					s.logger.Error("Failed to set pin state: ", "error", err)
				}
			}
		}
	}()

	return nil
}

func (s *gpioFlickerGpioFlicker) NewClientFromConn(ctx context.Context, conn rpc.ClientConn, remoteName string, name resource.Name, logger logging.Logger) (resource.Resource, error) {
	panic("not implemented")
}

func (s *gpioFlickerGpioFlicker) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	panic("not implemented")
}

func (s *gpioFlickerGpioFlicker) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
