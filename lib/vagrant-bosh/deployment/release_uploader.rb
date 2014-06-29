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

          # .dev_builds/ contains jobs/ and packages/ tgzs
          SyncedFolderRSync::RsyncHelper.rsync_single(@machine, @machine.ssh_info, {
            type:      :rsync, 
            hostpath:  File.join(host_dir, ".dev_builds"),
            guestpath: File.join(guest_dir, ".dev_builds"),
            disabled:  false,
          })

          # dev_releases/ contains dev release manifest files
          SyncedFolderRSync::RsyncHelper.rsync_single(@machine, @machine.ssh_info, {
            type:      :rsync, 
            hostpath:  File.join(host_dir, "dev_releases"),
            guestpath: File.join(guest_dir, "dev_releases"),
            disabled:  false,
          })
        end
      end

      #~
    end
  end
end
