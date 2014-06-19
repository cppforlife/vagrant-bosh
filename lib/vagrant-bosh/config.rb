require "vagrant"

module VagrantPlugins
  module VagrantBosh
    class Config < Vagrant.plugin("2", :config)
      attr_accessor :manifest

      def finalize!
      end

      def validate(machine)
      end
    end
  end
end
