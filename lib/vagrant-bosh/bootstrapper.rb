require "log4r"
require "json"

module VagrantPlugins
  module VagrantBosh
    class Bootstrapper
      def initialize(communicator, asset_uploader, base_dir, config, provisioner_tracker)
        @c = communicator
        @asset_uploader = asset_uploader
        @base_dir = base_dir
        @config = config
        @provisioner_tracker = provisioner_tracker
        @logger = Log4r::Logger.new("vagrant::provisioners::bosh::bootstrapper")

        @assets_path = File.join(@base_dir, "assets")
        @repos_path  = File.join(@base_dir, "repos")
        @manifest_path = File.join(@base_dir, "manifest.yml")
        @config_path   = File.join(@base_dir, "config.json")
      end

      def bootstrap
        @asset_uploader.sync(@assets_path)

        if @config.manifest
          @asset_uploader.upload_text(@config.manifest, @manifest_path)
        end

        config_json = JSON.dump(config_hash)
        @asset_uploader.upload_text(config_json, @config_path)

        # Provisioner is already uploaded when assets are synced
        provisioner_path = File.join(@assets_path, "provisioner")
        @c.chmod_x(provisioner_path)

        @c.sudo(
          "#{provisioner_path} -configPath=#{@config_path} 2> >(tee /tmp/provisioner.log >&2)",
          &@provisioner_tracker.method(:add_data)
        )
      end

      private

      def config_hash
        {
          assets_dir: @assets_path,
          repos_dir: @repos_path,

          blobstore: {
            provider: "local",
            options: {
              blobstore_path: File.join(@base_dir, "blobstore"),
            },
          },

          vm_provisioner: {
            full_stemcell_compatibility: @config.full_stemcell_compatibility,

            agent_provisioner: {
              infrastructure: @config.agent_infrastructure,
              platform:       @config.agent_platform,
              configuration:  @config.agent_configuration,

              mbus: "https://user:password@127.0.0.1:4321/agent",
            },
          },

          deployment_provisioner: {
            manifest_path: (@manifest_path if @config.manifest),
          },
        }
      end
    end
  end
end
