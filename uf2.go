//go:build !tinygo

package thing2

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/merliot/thing2/target"
)

// Random string to embed in UF2 we can search for later to locate params
const uf2Magic = "Call the Doctor!  Miss you Dan."

type Uf2Params struct {
	MagicStart   string
	Ssid         string
	Passphrase   string
	Id           string
	Model        string
	Name         string
	DeployParams template.HTML
	User         string
	Passwd       string
	DialURLs     string
	MagicEnd     string
}

func GenerateUf2s(dir string) error {
	for name, model := range Models {
		var proto = &Device{Model: name}
		proto.build(model.Maker)
		if err := proto.generateUf2s(dir); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) generateUf2s(dir string) error {
	for _, target := range target.TinyGoTargets(d.Targets) {
		if err := d.generateUf2(dir, target); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) generateUf2(dir, target string) error {

	// Create temp build directory
	temp, err := os.MkdirTemp("./", d.Model+"-")
	if err != nil {
		return err
	}

	if keepBuilds != "" {
		fmt.Println("DEBUG: Temporary build dir:", temp)
	} else {
		defer os.RemoveAll(temp)
	}

	var runnerGo = filepath.Join(temp, "runner.go")
	if err := d.genFile("device-runner-tinygo.tmpl", runnerGo, pageVars{}); err != nil {
		return err
	}

	// Build the uf2 file
	uf2Name := d.Model + "-" + target + ".uf2"
	output := filepath.Join(dir, uf2Name)
	cmd := exec.Command("tinygo", "build", "-target", target, "-o", output, "-stack-size", "8kb", "-size", "short", runnerGo)
	fmt.Println(cmd.String())
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}
	fmt.Println(string(stdoutStderr))

	return nil
}
