package sensehat

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/edaniels/golog"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/resource"
	"go.viam.com/utils"
)

var LPS25HModel = resource.NewModel("viamlabs", "i2c", "lps25h")

const (
	// lps25h address source -> https://pinout.xyz/pinout/sense_hat / https://www.st.com/resource/en/datasheet/lps25h.pdf
	defaultI2Caddr = 0x5c
	// TODO: Check if lps25h has a reset register
	lps25hRSTReg          = 0xE0 // Softreset Reg
	lps25hMeasurementsReg = 0x5c // TODO: find appropriate lps25h register

	// from here:  https://github.com/davebm1/c-sense-hat/blob/main/pressure.c
	DEV_ID       = 0x5c
	DEV_PATH     = "/dev/i2c-1"
	WHO_AM_I     = 0x0F
	CTRL_REG1    = 0x20
	CTRL_REG2    = 0x21
	PRESS_OUT_XL = 0x28
	PRESS_OUT_L  = 0x29
	PRESS_OUT_H  = 0x2A
	TEMP_OUT_L   = 0x2B
	TEMP_OUT_H   = 0x2C
)

// Config is used for converting config attributes.
type Config struct {
	Board   string `json:"board"`
	I2CBus  string `json:"i2c_bus"`
	I2cAddr int    `json:"i2c_addr,omitempty"`
}

// Validate ensures all parts of the config are valid.
func (conf *Config) Validate(path string) ([]string, error) {
	var deps []string
	if len(conf.Board) == 0 {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "board")
	}
	deps = append(deps, conf.Board)
	if len(conf.I2CBus) == 0 {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "i2c bus")
	}
	return deps, nil
}

func init() {
	resource.RegisterComponent(
		sensor.API,
		LPS25HModel,
		resource.Registration[sensor.Sensor, *Config]{
			Constructor: func(
				ctx context.Context,
				deps resource.Dependencies,
				conf resource.Config,
				logger golog.Logger,
				// TODO: Verify newConf as difers to Shawn's code
			) (sensor.Sensor, error) {
				newConf, err := resource.NativeConfig[*Config](conf)
				if err != nil {
					return nil, err
				}
				return newSensor(ctx, deps, conf.ResourceName(), newConf, logger)
			},
		},
	)
}

func newSensor(
	ctx context.Context,
	deps resource.Dependencies,
	name resource.Name,
	conf *Config,
	logger golog.Logger,
) (
	sensor.Sensor, error,
) {
	b, err := board.FromDependencies(deps, conf.Board)
	if err != nil {
		return nil, fmt.Errorf("lps25h init: failed to find board: %w", err)
	}
	localB, ok := b.(board.LocalBoard)
	if !ok {
		return nil, fmt.Errorf("board %s is not local", conf.Board)
	}
	i2cbus, ok := localB.I2CByName(conf.I2CBus)
	if !ok {
		return nil, fmt.Errorf("lps25h init: failed to find i2c bus %s", conf.I2CBus)
	}
	addr := conf.I2cAddr
	if addr == 0 {
		addr = defaultI2Caddr
		logger.Warn("using i2c address : 0x5c")
	}

	s := &lps25h{
		Named:    name.AsNamed(),
		logger:   logger,
		bus:      i2cbus,
		addr:     byte(addr),
		lastTemp: -999, // initialize to impossible temp
	}

	err = s.reset(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: Check if reset handling logic needs to be added from bme280 starting at line 158

	return s, nil
}

// lps25h is a sensor device
type lps25h struct {
	resource.Named
	resource.AlwaysRebuild
	resource.TriviallyCloseable

	mu     sync.Mutex
	logger golog.Logger

	bus         board.I2C
	addr        byte
	calibration map[string]int
	lastTemp    float64 // Store raw data from temp for humidity calculations

}

// TODO: Implement
func (s *lps25h) Readings(
	ctx context.Context,
	_ map[string]interface{},
) (map[string]interface{}, error) {

	handle, err := s.bus.OpenHandle(s.addr)
	if err != nil {
		s.logger.Errorf("can't open lps25h i2c %s", err)
		return nil, err
	}
	err = handle.Write(ctx, []byte{byte(lps25hMeasurementsReg)})
	if err != nil {
		s.logger.Debug("Failed to request temperature")
	}
	buffer, err := handle.Read(ctx, 8)
	if err != nil {
		return nil, err
	}
	if len(buffer) != 8 {
		return nil, errors.New("i2c read did not get 8 bytes")
	}

	return map[string]interface{}{
		"id": string(buffer),
	}, handle.Close()
}

func (s *lps25h) reset(ctx context.Context) error {
	handle, err := s.bus.OpenHandle(s.addr)
	if err != nil {
		return err
	}
	//err = handle.WriteByteData(ctx, lps25hRSTReg, 0xB6) // TODO: Be careful with writing!
	if err != nil {
		return err
	}
	return handle.Close()
}

func (s *lps25h) Reconfigure(
	_ context.Context,
	_ resource.Dependencies,
	conf resource.Config,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil
}
