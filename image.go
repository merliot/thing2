package thing2

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	keepBuilds = getenv("DEBUG_KEEP_BUILDS", "")
)

func (d *Device) genFile(template string, name string, pageVars pageVars) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	return d.renderPage(file, template, pageVars)
}

func (d *Device) buildLinuxImage(w http.ResponseWriter, r *http.Request, dir string,
	envs []string, target string) error {

	var user = getenv("USER", "")
	var passwd = getenv("PASSWD", "")
	var port = r.URL.Query().Get("port")
	var url = r.Header["Hx-Current-Url"][0]
	var dialurls = strings.Replace(url, "http", "ws", 1) + "ws"
	var service = d.Model + "-" + d.Id

	// Generate runner.go from device-runner.tmpl
	var runnerGo = filepath.Join(dir, "runner.go")
	if err := d.genFile("device-runner.tmpl", runnerGo, pageVars{
		"progenitive": d.cfg.Test(FlagProgenitive),
		"user":        user,
		"passwd":      passwd,
		"dialurls":    dialurls,
		"port":        port,
	}); err != nil {
		return err
	}

	// Generate installer.go from device-installer.tmpl
	var installerGo = filepath.Join(dir, "installer.go")
	if err := d.genFile("device-installer.tmpl", installerGo, pageVars{
		"service": service,
	}); err != nil {
		return err
	}

	// Generate {{service}}.service from device-service.tmpl
	var output = filepath.Join(dir, service+".service")
	if err := d.genFile("device-service.tmpl", output, pageVars{
		"service": service,
	}); err != nil {
		return err
	}

	// Generate {{service}}.conf from device-conf.tmpl
	output = filepath.Join(dir, service+".conf")
	if err := d.genFile("device-conf.tmpl", output, pageVars{
		"service": service,
	}); err != nil {
		return err
	}

	// Build runner binary

	// substitute "-" for "_" in target, ala TinyGo, when using as tag
	var tag = strings.Replace(target, "-", "_", -1)
	var binary = filepath.Join(dir, service)

	cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", binary, "-tags", tag, runnerGo)
	fmt.Println(cmd.String())
	cmd.Env = append(cmd.Environ(), envs...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	// Build installer

	var installer = filepath.Join(dir, service+"-installer")

	cmd = exec.Command("go", "build", "-ldflags", "-s -w", "-o", installer, installerGo)
	fmt.Println(cmd.String())
	cmd.Env = append(cmd.Environ(), envs...)
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	println("done")

	return nil
}

func (d *Device) buildImage(w http.ResponseWriter, r *http.Request) error {

	// Create temp build directory
	dir, err := os.MkdirTemp("./", d.Id+"-")
	if err != nil {
		return err
	}

	if keepBuilds != "" {
		fmt.Println("DEBUG: Temporary build dir:", dir)
	} else {
		defer os.RemoveAll(dir)
	}

	target := r.URL.Query().Get("target")

	switch target {
	case "demo", "x86-64":
		envs := []string{"CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64"}
		return d.buildLinuxImage(w, r, dir, envs, target)
	case "rpi":
		// TODO: do we want more targets for GOARM=7|8?
		envs := []string{"CGO_ENABLED=0", "GOOS=linux", "GOARCH=arm", "GOARM=5"}
		return d.buildLinuxImage(w, r, dir, envs, target)
		/*
			case "nano-rp2040", "wioterminal", "pyportal":
				//return d.deployTinyGo(dir, values, envs, templates, w, r)
				return d.deployTinyGoUF2(dir, values, envs, templates, w, r)
		*/
	default:
		return fmt.Errorf("Target '%s' not supported", target)
	}

	return nil
}

func (d *Device) downloadImage(w http.ResponseWriter, r *http.Request) {

	if d.Locked {
		http.Error(w, "Refusing to download, device is locked", http.StatusLocked)
		return
	}

	// The r.URL values are passed in from the download <form>.  These
	// values are the proposed new device config, and should decode into
	// the device.  If accepted, the device is updated and the config is
	// stored in DeployParams.

	changed, err := d.formConfig(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	// Built it!

	if err := d.buildImage(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	// If the device config has changed, kick the downlink device offline.
	// It will try to reconnect, but fail, because the DeployParams now
	// don't match this (uplink) device.  Once the downlink device is
	// updated (with the image we created above) the downlink device
	// will connect.

	if changed {
		dirty.Store(true)
		downlinkClose(d.Id)
	}
}
