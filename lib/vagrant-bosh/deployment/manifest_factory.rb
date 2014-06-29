require "vagrant-bosh/deployment/manifest"

module VagrantPlugins
  module VagrantBosh
    module Deployment
      class ManifestFactory
        def initialize(uploadable_release_factory, ui)
          @uploadable_release_factory = uploadable_release_factory
          @ui = ui
        end

        def new_manifest(manifest)
          if manifest.empty?
            EmptyManifest.new
          else
            Manifest.new(manifest, @uploadable_release_factory, @ui)
          end
        end
      end

      class EmptyManifest
        def resolve_releases; end
        def as_string;    ""; end
      end

      #~
    end
  end
end
