begin
  require "vagrant"
rescue LoadError
  raise "The BOSH Release Vagrant plugin must be run within Vagrant."
end

module VagrantPlugins
  module VagrantBosh
    class Plugin < Vagrant.plugin("2")
      name "bosh"

      description "Provisions virtual machines via BOSH deployment manifest"

      config(:bosh, :provisioner) do
        require File.expand_path("../config", __FILE__)
        Config
      end

      provisioner(:bosh) do
        require File.expand_path("../provisioner", __FILE__)
        Provisioner
      end
    end
  end
end
