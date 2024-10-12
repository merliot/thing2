package thing2

import (
	"fmt"
	"html/template"
	"math"
	"net/url"
	"sync"
	"time"
)

type Devicer interface {
	GetConfig() Config
	GetHandlers() Handlers
	Setup() error
	DemoSetup() error
	Poll(*Packet)
	DemoPoll(*Packet)
}

type Device struct {
	Id           string
	Model        string
	Name         string
	Children     []string
	DeployParams template.HTML
	flags        `json:"-"`
	Config       `json:"-"`
	Devicer      `json:"-"`
	Handlers     `json:"-"`
	sync.RWMutex `json:"-"`
	deviceOS
	stopChan chan struct{}
}

func (d *Device) build(maker Maker) error {

	d.Devicer = maker()
	d.Config = d.GetConfig()
	d.Handlers = d.GetHandlers()
	d.flags = d.Config.Flags

	if runningDemo {
		d.Set(flagDemo | flagOnline | flagMetal)
	}

	// Bracket poll period: [1..forever) seconds
	if d.PollPeriod == 0 {
		d.PollPeriod = time.Duration(math.MaxInt64)
	} else if d.PollPeriod < time.Second {
		d.PollPeriod = time.Second
	}

	// Configure the device using DeployParams
	_, err := d.formConfig(string(d.DeployParams))
	if err != nil {
		fmt.Println("Error configuring device using DeployParams:", err, d)
	}

	d.stopChan = make(chan struct{})

	return d.buildOS()
}

func (d *Device) formConfig(rawQuery string) (changed bool, err error) {

	// rawQuery is the proposed new DeployParams
	proposedParams, err := url.QueryUnescape(rawQuery)
	if err != nil {
		return false, err
	}
	values, err := url.ParseQuery(proposedParams)
	if err != nil {
		return false, err
	}

	d.Lock()
	defer d.Unlock()

	//	fmt.Println("Proposed DeployParams:", proposedParams)

	// Form-decode these values into the device to configure the device
	if err := decoder.Decode(d.State, values); err != nil {
		return false, err
	}

	if proposedParams == string(d.DeployParams) {
		// No change
		return false, nil
	}

	// Save changes.  Store DeployParams unescaped.
	d.DeployParams = template.HTML(proposedParams)
	return true, nil
}
