package main

import (
	"fmt"
	"os"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/hashicorp/packer/template/interpolate"
	"bytes"
	_ "runtime"
)

// SrvSpeConfig holds the config data coming in from the packer template
type SrvSpecConfig struct {

	// provisioner version
	Version string

	// command to retrieve remote server hostname /ip
	CMD string

	// An array of tests to run.
	TestSpecsDir string `mapstructure:"test_specs_dir"`

	// User For SSH
	SshUser string `mapstructure:"ssh_user"`

	ctx interpolate.Context
}

// Provisioner implements a packer Provisioner
type Provisioner struct {
	config SrvSpecConfig
}

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterProvisioner(new(Provisioner))
	server.Serve()
}

// Prepare gets the ServerSpec Provisioner ready to run
func (p *Provisioner) Prepare(raws ...interface{}) error {
	var errs *packer.MultiError
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.Version == "" {
		p.config.Version = "0.0.1"
	}

	if p.config.CMD == "" {
		p.config.CMD = "ip -4 route get 1|head -1|awk '{print $NF}'"
	}

	if p.config.TestSpecsDir == "" {
		p.config.TestSpecsDir = "serverspec"
	}

	if p.config.SshUser == "" {
		p.config.SshUser = "centos"
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

// Provision runs the ServerSpec Provisioner
func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	ui.Say("Running The ServerSpec Provisioner")

	ui.Say("Getting Remote Host IP")
	hostInfo, err := p.getRemoteHostIP(ui, comm)
	if err != nil {
		return fmt.Errorf("Error Getting Remote Host IP Details: %s", err)
	}
	ui.Message(fmt.Sprintf("Remote Host IP: %s", hostInfo))

	ui.Say("Running ServerSpec tests - Hopefully Soon...")
	/*
	if err := p.runSrvSpec(ui, comm, hostInfo); err != nil {
		return fmt.Errorf("Error running ServerSpec: %s", err)
	}
	*/
	return nil
}

// Cancel just exists when provision is cancelled
func (p *Provisioner) Cancel() {
	os.Exit(0)
}

// Get Remote Host From Remote Machine
func (p *Provisioner) getRemoteHostIP(ui packer.Ui, comm packer.Communicator) (string, error) {
	var stdout bytes.Buffer
	// create a file and write output to that file.
	file, err := os.Create("remote-data.txt")
	if err != nil {
		return "", fmt.Errorf("Error :: Could not create file: %s", err)
	}
	defer file.Close()

	ui.Message(fmt.Sprintf("Asking Remote Host for: %s", p.config.CMD))
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("'%s'", p.config.CMD),
		Stdout: &stdout,
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return "", err
	}
	if cmd.ExitStatus != 0 {
		return "", fmt.Errorf("non-zero exit status")
	}
	fmt.Fprintf(file, stdout.String())
	return stdout.String(), nil
}

// runSrvSpec runs the ServerSpec tests
/*
func (p *Provisioner) runSrvSpec(ui packer.Ui, _ packer.Communicator, hostInfo string) error {

	//commandRun :=
	var execCmd string[]
	commandRun := fmt.Sprintf("export SSH_USER=%s && cd %s| rake spec TARGET_HOST=%s",
		p.config.SshUser, p.config.TestSpecsDir, hostInfo)

	execCmd = string[] {
	"bin/bash",
	"-c",
	commandRun
	}

	comm := &Communicator.{
	Ctx:            p.config.ctx,
		ExecuteCommand: execCmd
	}
	return nil
}
*/
