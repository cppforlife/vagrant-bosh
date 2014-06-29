require "vagrant"

module VagrantPlugins
  module VagrantBosh
    module Errors
      class BoshReleaseError < StandardError; end
    end
  end
end
