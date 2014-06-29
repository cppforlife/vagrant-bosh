require "log4r"
require "json"

module VagrantPlugins
  module VagrantBosh
    class Bootstrapper
      def initialize(communicator, config, asset_uploader, provisioner_tracker, manifest_factory)
        @c = communicator
        @config = config
        @asset_uploader = asset_uploader
        @provisioner_tracker = provisioner_tracker
        @manifest_factory = manifest_factory

        @logger = Log4r::Logger.new("vagrant::provisioners::bosh::bootstrapper")
      end

      def bootstrap
        @asset_uploader.sync(@config.assets_dir)

        if @config.manifest
          manifest = @manifest_factory.new_manifest(@config.manifest)
          manifest.resolve_releases
          @asset_uploader.upload_text(manifest.as_string, @config.manifest_path)
        end

        config_json = JSON.dump(config_hash)
        @asset_uploader.upload_text(config_json, @config.config_path)

        run_provisioner
      end

      private

      def run_provisioner
        # Provisioner is already uploaded when assets are synced
        provisioner_path = File.join(@config.assets_dir, "provisioner")
        @c.chmod_x(provisioner_path)

        @c.sudo(
          "#{provisioner_path} -configPath=#{@config.config_path} 2> >(tee /tmp/provisioner.log >&2)",
          &@provisioner_tracker.method(:add_data)
        )
      end

      def config_hash
        {
          assets_dir: @config.assets_dir,
          repos_dir: @config.repos_dir,

          blobstore: {
            provider: "local",
            options: {
              blobstore_path: @config.local_blobstore_dir,
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
            manifest_path: (@config.manifest_path if @config.manifest),
          },
        }
      end
    end
  end
end
