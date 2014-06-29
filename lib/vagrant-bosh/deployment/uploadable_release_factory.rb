require "vagrant-bosh/deployment/uploadable_release"

module VagrantPlugins
  module VagrantBosh
    module Deployment
      class UploadableReleaseFactory
        def initialize(guest_root_dir, release_uploader, ui)
          @guest_root_dir = guest_root_dir
          @release_uploader = release_uploader
          @ui = ui
        end

        def new_uploadable_release(name, version, host_dir)
          UplodableRelease.new(
            name,
            version,
            host_dir, 
            @guest_root_dir, 
            @release_uploader, 
            @ui,
          )
        end
      end

      #~
    end
  end
end
