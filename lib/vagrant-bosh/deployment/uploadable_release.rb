require "log4r"
require "vagrant/util/subprocess"
require "vagrant-bosh/errors"

require Vagrant.source_root.join("plugins/synced_folders/rsync/helper")

module VagrantPlugins
  module VagrantBosh
    module Deployment
      # UplodableRelease represents a release 
      # that *can be* synced to a guest FS location.
      class UplodableRelease
        def initialize(name, version, host_dir, guest_root_dir, release_uploader, ui)
          @name = name
          @version = version
          @host_dir = host_dir
          @release_uploader = release_uploader

          @ui = ui.for(:deployment, :uploadable_release)
          @logger = Log4r::Logger.new("vagrant::provisioners::bosh::deployment::uploadable_release")

          @guest_dir = File.join(guest_root_dir, name)
        end

        def upload
          version = @version == "latest" ? create_release : @version

          # Sync either existing or newly-created release
          @release_uploader.sync(@host_dir, @guest_dir)

          UploadedRelease.new(@name, version, @guest_dir)
        end

        private

        def create_release
          result = @ui.timed_msg(:create_release, name: @name) do
            # Without clearing out environment Vagrant ruby env will be inherited
            Vagrant::Util::Subprocess.execute(
              "env",  "-i", "HOME=#{ENV["HOME"]}", "TERM=#{ENV["TERM"]}",
              "bash", "-l", "-c",
              "bosh -n create release --force",
              {workdir: @host_dir},
            )
          end

          if result.exit_code != 0
            error_msg = @ui.msg_string(:create_release_error, {
              name: @name,
              stdout: result.stdout,
              stderr: result.stderr,
            })
            raise VagrantPlugins::VagrantBosh::Errors::BoshReleaseError, error_msg
          end

          if result.stdout =~ /^Release version: (.+)$/
            version = $1
          else
            error_msg = @ui.msg_string(:missing_release_version_error, {
              name: @name,
              stdout: result.stdout,
              stderr: result.stderr,
            })
            raise VagrantPlugins::VagrantBosh::Errors::BoshReleaseError, error_msg
          end

          version
        end
      end

      # UploadedRelease represents a release 
      # that *was* synced to a guest FS location.
      class UploadedRelease < Struct.new(:name, :version, :guest_dir)
        def as_hash
          {"name" => name, "version" => version, "url" => "dir://#{guest_dir}"}
        end
      end

      #~
    end
  end
end
