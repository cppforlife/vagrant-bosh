require "vagrant"

module VagrantPlugins
  module VagrantBosh
    module Errors
      class BoshReleaseError < Vagrant::Errors::VagrantError
        error_namespace("bosh")
      end
    end
  end
end
