package thing2

import (
	"fmt"
	"net/http"
	"os"
)

var (
	keepBuilds = getEnv("DEBUG_KEEP_BUILDS", "")
)

/*
func (d *Device) genFile(templates *template.Template, template string, name string,
	values map[string]string) error {

	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := templates.Lookup(template)
	if tmpl == nil {
		return fmt.Errorf("Template '%s' not found", template)
	}

	return tmpl.Execute(file, values)
}
*/

func (d *Device) buildLinuxImage(w http.ResponseWriter, r *http.Request, dir string, envs []string) error {
	/*

		// Generate build.go from server.tmpl
		if err := d.genFile(templates, "server.tmpl", filepath.Join(dir, "build.go"), values); err != nil {
			return err
		}

		// Generate installer.go from installer.tmpl
		if err := genFile(templates, "installer.tmpl", filepath.Join(dir, "installer.go"), values); err != nil {
			return err
		}

		// Generate model.service from service.tmpl
		if err := genFile(templates, "service.tmpl", filepath.Join(dir, d.Model+".service"), values); err != nil {
			return err
		}

		// Generate model.conf from log.tmpl
		if err := genFile(templates, "log.tmpl", filepath.Join(dir, d.Model+".conf"), values); err != nil {
			return err
		}

		// Build build.go -> model (binary)

		// substitute "-" for "_" in target, ala TinyGo, when using as tag
		target := strings.Replace(values["target"], "-", "_", -1)

		cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", filepath.Join(dir, d.Model),
			"-tags", target, filepath.Join(dir, "build.go"))
		fmt.Println(cmd.String())
		cmd.Env = append(cmd.Environ(), envs...)
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%w: %s", err, stdoutStderr)
		}

		// Build installer and serve as download-able file

		installer := d.Id + "-installer"
		cmd = exec.Command("go", "build", "-ldflags", "-s -w", "-o", filepath.Join(dir, installer), filepath.Join(dir, "installer.go"))
		fmt.Println(cmd.String())
		cmd.Env = append(cmd.Environ(), envs...)
		stdoutStderr, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%w: %s", err, stdoutStderr)
		}

		return d.serveFile(dir, installer, w, r)
	*/
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
		return d.buildLinuxImage(w, r, dir, envs)
	case "rpi":
		// TODO: do we want more targets for GOARM=7|8?
		envs := []string{"CGO_ENABLED=0", "GOOS=linux", "GOARCH=arm", "GOARM=5"}
		return d.buildLinuxImage(w, r, dir, envs)
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
		downlinkClose(d.Id)
	}
}
