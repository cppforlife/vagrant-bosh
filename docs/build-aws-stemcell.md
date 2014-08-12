## Building AWS BOSH stemcell with BOSH provisioner

!!! Stemcells produced using this method should NOT be used in production !!!

To quickly create BOSH stemcells for AWS:

1. Configure AWS Vagrant provider with your AWS account settings. For example:

```
# AWS provider is configured to use an AMI
config.vm.box = "dummy"

config.vm.provider "aws" do |aws, override|
  aws.access_key_id = "AKIxxx"
  aws.secret_access_key = "xxx"

  aws.keypair_name = "my-aws-key-pair-name"
  aws.security_groups = ['my-aws-sec-group']

  aws.tags = { "Name" => "vagrant-bosh-stemcell" }

  aws.block_device_mapping = [{
    :DeviceName => "/dev/sda1",
    "Ebs.VolumeSize" => 8
  }]

  # https://cloud-images.ubuntu.com/trusty/current/
  aws.ami = "ami-018c9568"

  override.ssh.username = "ubuntu"
  override.ssh.private_key_path = "/Users/some-user/.ssh/my-aws-private-key"
end
```

2. Configure BOSH provisioner with following settings:

```
config.vm.provision "bosh" do |c|
  c.manifest = nil

  # Some BOSH releases (e.g. cf-release) depend on more pre-installed packages
  c.full_stemcell_compatibility = true

  c.agent_infrastructure = "aws"
  c.agent_platform = "ubuntu"
  c.agent_configuration = {}
end
```

3. Remove `cloud-init` package included by default in Ubuntu AMIs
   to avoid conflicts with BOSH Agent auto configuration:

```
config.vm.provision "shell", inline: "apt-get -y purge --auto-remove cloud-init"
config.vm.provision "shell", inline: "echo 'LABEL=cloudimg-rootfs /  ext4 defaults  0 0' > /etc/fstab"
```

4. `vagrant up --provider aws`

5. Once VM is provisioned, in AWS Console right-click on the VM and select `Cerate image` to create an AMI.
   Optionally make AMI public by changing its permissions if stemcell will be used from a different AWS account.

6. Unpack one of the officially published `light-bosh` stemcells and
   update `stemcell.MF` with new AMI reference, then repack.

7. Upload your new light stemcell to a BOSH Director and use it in your deployment.
