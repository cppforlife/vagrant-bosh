require "log4r"
require "vagrant-bosh/ui"
require "vagrant-bosh/communicator"
require "vagrant-bosh/asset_uploader"
require "vagrant-bosh/bootstrapper"
require "vagrant-bosh/provisioner_tracker"

module VagrantPlugins
  module VagrantBosh
    class Provisioner < Vagrant.plugin("2", :provisioner)
      def initialize(machine, config)
        super

        machine_ui = Ui.for_machine(machine)

        communicator = Communicator.new(machine, machine_ui)

        asset_uploader = AssetUploader.new(
          communicator,
          machine_ui,
          File.absolute_path("../assets", __FILE__),
        )

        provisioner_tracker = ProvisionerTracker.new(machine_ui)

        @bootstrapper = Bootstrapper.new(
          communicator,
          asset_uploader,
          "/opt/vagrant-bosh",
          config.manifest,
          provisioner_tracker,
        )

        logger = Log4r::Logger.new("vagrant::provisioners::bosh")
      end

      def configure(root_config)
        # Nothing to modify or save of config
      end

      def provision
        @bootstrapper.bootstrap
      end
    end
  end
end
