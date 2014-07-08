require "log4r"

require Vagrant.source_root.join("plugins/synced_folders/rsync/helper")

module VagrantPlugins
  module VagrantBosh
    module Deployment
      class ReleaseUploader
        def initialize(machine, ui)
          @machine = machine

          @ui = ui.for(:deployment, :release_uploader)
          @logger = Log4r::Logger.new("vagrant::provisioners::bosh::deployment::release_uploader")
        end

        def sync(host_dir, guest_dir)
          # RsyncHelper uses @machine.ui internally

          dir_names = [
            # .dev_builds/ and .final_builds/ contain jobs/ and packages/ tgzs
            ".dev_builds",
            ".final_builds",

            # dev_releases/ contains dev release manifest files
            "dev_releases",            
          ]

          dir_names.each do |dir_name|
            SyncedFolderRSync::RsyncHelper.rsync_single(@machine, @machine.ssh_info, {
              type:      :rsync, 
              hostpath:  File.join(host_dir, dir_name),
              guestpath: File.join(guest_dir, dir_name),
              disabled:  false,
            })
          end
        end
      end

      #~
    end
  end
end
