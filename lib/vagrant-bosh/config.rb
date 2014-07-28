require "vagrant"

module VagrantPlugins
  module VagrantBosh
    class Config < Vagrant.plugin("2", :config)
      attr_accessor :base_dir

      attr_reader :assets_dir, :repos_dir

      attr_reader :manifest_path, :config_path

      attr_reader :local_blobstore_dir, :synced_releases_dir

      # Manifest holds full BOSH deployment manifest as a string.
      attr_accessor :manifest

      # Full stemcell compatibility forces provisioner to install all 
      # (not just minimum) dependencies usually found on a stemcell.
      attr_accessor :full_stemcell_compatibility

      attr_accessor :agent_infrastructure, :agent_platform, :agent_configuration

      # Command to run to create a BOSH release on the host.
      # BOSH release has a rake task to create a dev release 
      # because it creates locally versioned gems.
      attr_accessor :create_release_cmd

      def initialize(*args)
        super
        @base_dir = "/opt/bosh-provisioner"
      end

      def finalize!
        @assets_dir = File.join(@base_dir, "assets")
        @repos_dir  = File.join(@base_dir, "repos")

        @manifest_path = File.join(@base_dir, "manifest.yml")
        @config_path   = File.join(@base_dir, "config.json")

        @local_blobstore_dir = File.join(@base_dir, "blobstore")
        @synced_releases_dir = File.join(@base_dir, "synced-releases")

        @full_stemcell_compatibility = !!@full_stemcell_compatibility

        @create_release_cmd ||= "ruby -v; bosh -n create release --force"
      end

      def validate(machine)
      end
    end
  end
end
