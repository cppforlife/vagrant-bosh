require "vagrant"

module VagrantPlugins
  module VagrantBosh
    class Config < Vagrant.plugin("2", :config)
      # Manifest holds full BOSH deployment manifest as a string.
      attr_accessor :manifest

      # Full stemcell compatibility forces provisioner to install all 
      # (not just minimum) dependencies usually found on a stemcell.
      attr_accessor :full_stemcell_compatibility

      attr_accessor :agent_infrastructure, :agent_platform, :agent_configuration

      def finalize!
        self.full_stemcell_compatibility = !!self.full_stemcell_compatibility
      end

      def validate(machine)
      end
    end
  end
end
