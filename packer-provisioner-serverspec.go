package main

import (
	"fmt"
	"os"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/hashicorp/packer/template/interpolate"
	"bytes"
	"runtime"
	"errors"
	shellLocal "github.com/hashicorp/packer/provisioner/shell-local"
)

// SrvSpeConfig holds the config data coming in from the packer template
type SrvSpecConfig struct {

	common.PackerConfig `mapstructure:",squash"`

	// provisioner version
	Version string

	// command to retrieve remote server hostname /ip
	CMD string `mapstructure:"remote_host_command"`

	// The location of the serverspec tests
	TestSpecsDir string `mapstructure:"test_specs_dir"`

	// User For SSH
	SshUser string `mapstructure:"ssh_user"`

	// The os type (windows or linux) this is only applied if SrvSpecCommand isnt specified
	OSType string `mapstructure:"os_type"`

	// Should the actual serverspec command be run - or do you just want the ip of the remote host ?
	RunSrvSpec bool `mapstructure:"run_serverspec"`

	// ExecuteCommand is the command used to execute the command.
	ExecuteCommand []string `mapstructure:"execute_command"`

	// The Actual ServerSpec command to exist, can be overriden,
	// if left as-is it will use osType and run different commands
	// depending if it is windows (untested) or linux
	SrvSpecCommand string `mapstructure:"serverspec_command"`

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
			Exclude: []string{
				"execute_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.Version == "" {
		p.config.Version = "0.0.1"
	}

	if p.config.CMD == "" {
		p.config.CMD = "sudo hostname -i"
	}

	if p.config.TestSpecsDir == "" {
		p.config.TestSpecsDir = "serverspec"
	}

	if p.config.SshUser == "" {
		p.config.SshUser = "centos"
	}

	if len(p.config.ExecuteCommand) == 0 {
		if runtime.GOOS == "windows" {
			p.config.ExecuteCommand = []string{
				"cmd",
				"/C",
				"{{.Command}}",
			}
		} else {
			p.config.ExecuteCommand = []string{
				"/bin/sh",
				"-c",
				"{{.Command}}",
			}
		}
	}

	if len(p.config.ExecuteCommand) == 0 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("execute_command must not be empty"))
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

	if p.config.RunSrvSpec == true {
		ui.Say("Running ServerSpec tests")
		if err := p.runSrvSpec(ui, hostInfo); err != nil {
			return fmt.Errorf("Error running ServerSpec: %s", err)
		}

	}

	return nil
}

// Cancel just exists when provision is cancelled
func (p *Provisioner) Cancel() {
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
		Command: fmt.Sprintf("%s", p.config.CMD),
		Stdout: &stdout,
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return "", err
	}
	cmd.Wait()
	if cmd.ExitStatus != 0 {
		return "", fmt.Errorf("non-zero exit status")
	}

	fmt.Fprintf(file, stdout.String())
	return stdout.String(), nil
}

// runSrvSpec runs the ServerSpec tests
func (p *Provisioner) runSrvSpec(ui packer.Ui, hostInfo string) error {
	var stdout bytes.Buffer

	if p.config.SrvSpecCommand == "" {
		if p.config.OSType == "windows" {
			p.config.SrvSpecCommand = fmt.Sprintf("cd %s && rake spec",
				p.config.TestSpecsDir)
		} else {
			p.config.SrvSpecCommand = fmt.Sprintf("cd %s && SSH_USER=%s rake spec TARGET_HOST=%s",
				p.config.TestSpecsDir,
				p.config.SshUser,
				hostInfo)
		}
	}

	comm := &shellLocal.Communicator{
		Ctx:            p.config.ctx,
		ExecuteCommand: p.config.ExecuteCommand,
	}

	cmd := &packer.RemoteCmd{
		Command: p.config.SrvSpecCommand,
		Stdout: &stdout,
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return fmt.Errorf( "Error executing: %s\n",
			p.config.SrvSpecCommand)
	}
	fmt.Sprintf("Output: %s", stdout.String())
	cmd.Wait()
	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Error :: Code: %d command: %s",
			cmd.ExitStatus,
			p.config.SrvSpecCommand)
	}
	return nil
}

