//go:build rpi

package relay

import (
	"fmt"
	"strconv"

	"github.com/merliot/thing2/target"
	"gobot.io/x/gobot/v2/drivers/gpio"
)

type Relay struct {
	Name   string
	Gpio   string
	State  bool
	driver *gpio.RelayDriver
}

func (r *Relay) Setup() error {
	if pin, ok := target.Pin(r.Gpio); ok {
		spin := strconv.Itoa(int(pin))
		fmt.Println(r.Gpio, pin, spin)
		r.driver = gpio.NewRelayDriver(target.Adaptor, spin)
		fmt.Println("Setup r.driver", r, r.driver)
		r.driver.Start()
		r.driver.Off()
		return nil
	}
	return fmt.Errorf("No pin for GPIO %s", r.Gpio)
}

func (r *Relay) Set(state bool) {
	fmt.Println("Set r.driver", r.driver)
	if r.driver != nil {
		r.State = state
		if state {
			fmt.Println(state, r.driver.State())
			fmt.Println("ON")
			r.driver.On()
			fmt.Println(state, r.driver.State())
		} else {
			fmt.Println(state, r.driver.State())
			fmt.Println("OFF")
			r.driver.Off()
			fmt.Println(state, r.driver.State())
		}
	}
}
