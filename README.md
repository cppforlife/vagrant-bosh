## BOSH provisioner for Vagrant

BOSH provisioner allows to provision guest VM by specifying regular BOSH deployment manifest.


### Usage

1. Add new VM provision section to your `Vagrantfile`. For example:

```
Vagrant.configure("2") do |config|
  config.vm.box = "precise64"
  config.vm.box_url = "http://files.vagrantup.com/precise64.box"

  config.vm.network "forwarded_port", guest: 25555, host: 25555 # for BOSH Director API

  config.vm.provider "virtualbox" do |v|
    v.memory = 4096
    v.cpus = 2
  end

  config.vm.provision "bosh" do |c|
    # use cat or just inline full deployment manifest
    c.manifest = `cat somewhere/example-bosh-manifest.yml`
  end
end
```

2. Create a deployment manifest and specify it via `c.manifest` attribute.
   See `dev/example-bosh-manifest.yml` for an example deployment manifest used to deploy BOSH Director.

3. Run `vagrant provision` to provision guest VM
   (DEBUG=1 environment variable will trigger live verbose output).


### Deployment manifest gotchas

- It must specify release source(s) via `url` key in the `releases` section.

- It must have exactly one deployment job; however, deployment job
  can be made up from multiple job templates that come from multiple releases.

- It does not support `static` network type though `dynamic` network type is supported
  (Network configuration should be done via standard Vagrant configuration DSL).

- It does not support stemcell specification because guest VM OS is picked via `config.vm.box` directive.


### Planned

- Speed up apply step (Monit is sluggish)

- Packer Provisioner API wrapper

- Support non-Ubuntu vagrant boxes (currently provisioner uses `apt-get` to bootstrap deps)

- Support non-tgz release source URLs


### Contributing

```
git submodule update --recursive --init

# Provisioner uses BOSH Agent (Go) libraries (e.g. Blobstore, CmdRunner)
export GOPATH=$PWD/go:$PWD/bosh/go_agent

go/bin/test
go/bin/build-linux-amd64

# Spin up development Vagrant box with lib/ acting as BOSH provisioner
( cd dev/ && vagrant up )
```
