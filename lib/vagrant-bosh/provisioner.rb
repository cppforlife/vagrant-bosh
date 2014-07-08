require "log4r"
require "vagrant-bosh/ui"
require "vagrant-bosh/communicator"
require "vagrant-bosh/asset_uploader"
require "vagrant-bosh/bootstrapper"
require "vagrant-bosh/provisioner_tracker"
require "vagrant-bosh/deployment/release_uploader"
require "vagrant-bosh/deployment/uploadable_release_factory"
require "vagrant-bosh/deployment/manifest_factory"

module VagrantPlugins
  module VagrantBosh
    class Provisioner < Vagrant.plugin("2", :provisioner)
      def initialize(machine, config)
        super

        machine_ui = Ui.for_machine(machine)

        communicator = Communicator.new(machine, machine_ui)

        asset_uploader = AssetUploader.new(
          communicator,
          File.absolute_path("../assets", __FILE__),
          machine_ui,
        )

        provisioner_tracker = ProvisionerTracker.new(machine_ui)

        release_uploader = Deployment::ReleaseUploader.new(
          machine, 
          machine_ui,
        )

        uploadable_release_factory = Deployment::UploadableReleaseFactory.new(
          config.synced_releases_dir,
          release_uploader,
          config.create_release_cmd,
          machine_ui,
        )

        manifest_factory = Deployment::ManifestFactory.new(
          uploadable_release_factory,
          machine_ui,
        )

        @bootstrapper = Bootstrapper.new(
          communicator, 
          config, 
          asset_uploader, 
          provisioner_tracker,
          manifest_factory, 
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
