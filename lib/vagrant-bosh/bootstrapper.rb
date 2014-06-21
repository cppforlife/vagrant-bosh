require "log4r"
require "json"

module VagrantPlugins
  module VagrantBosh
    class Bootstrapper
      def initialize(communicator, asset_uploader, base_dir, manifest, provisioner_tracker)
        @c = communicator
        @asset_uploader = asset_uploader
        @base_dir = base_dir
        @manifest = manifest
        @provisioner_tracker = provisioner_tracker
        @logger = Log4r::Logger.new("vagrant::provisioners::bosh::bootstrapper")

        @assets_path = File.join(@base_dir, "assets")
        @repos_path  = File.join(@base_dir, "repos")
        @manifest_path = File.join(@base_dir, "manifest.yml")
        @config_path   = File.join(@base_dir, "config.json")
      end

      def bootstrap
        @asset_uploader.sync(@assets_path)

        @asset_uploader.upload_text(@manifest, @manifest_path)

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
          manifest_path: @manifest_path,
          assets_dir: @assets_path,
          repos_dir: @repos_path,

          mbus: "https://user:password@127.0.0.1:4321/agent",

          blobstore: {
            provider: "local",
            options: {
              blobstore_path: File.join(@base_dir, "blobstore"),
            },
          },
        }
      end
    end
  end
end
