# packer-provisioner-serverspec

Run Serverspec as a packer provisioner step.

This has only been tested on linux, but should also work on windows.

## Methods

This can be used purely to get the remote provisioned ip address, especially useful if building aws ami images.
The provisioners default behaviour is to only run a query on the provisioned machine to get the ip address allocated,
pretty much all of the behaviour can be overridden.

Alternately this can be used to run serverspec. The configuration parameters are as follows:

```markdown
remote_host_command: (string) [default: "sudo hostname -i"]
command to retrieve remote server hostname / ip


test_specs_dir: (string) [default: "serverspec"]
The location of the serverspec tests (defaults to serverspec/



ssh_user: (string) [default: "centos"]
User For SSH Remote Connection



os_type: (string) [default: "linux"]
The os type (windows or linux),
this is only applied if serverspec_command isnt used



run_serverspec: (boolean) [default: true]
Should the actual serverspec command be run?



serverspec_command:
The ServerSpec command to run - ideally shouldnt be modified, but can be if required.
if left as-is it will use osType and try to run different commands depending on if it is
windows (untested) or linux
```

## Example

```json
"provisioners" : [
  {
    "type": "serverspec",
    "run_serverspec": true,
    "remote_host_command": "curl http://169.254.169.254/latest/meta-data/local-ipv4",
    "ssh_user": "ubuntu",
    "os_type": "linux"
  }
]
```


## Credits
This uses code directly from packer, it reuses shell-local and has parts taken from the shell-local provisioner.

This is based off of the work done for [packer-provisioner-goss](https://github.com/YaleUniversity/packer-provisioner-goss).


## License

### MIT

Copyright 2018 ev9.io

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
