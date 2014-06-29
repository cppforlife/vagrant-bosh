require "log4r"
require "yaml"

module VagrantPlugins
  module VagrantBosh
    module Deployment
      class Manifest
        def initialize(manifest, uploadable_release_factory, ui)
          @manifest = manifest
          @uploadable_release_factory = uploadable_release_factory

          @ui = ui.for(:deployment, :manifest)
          @logger = Log4r::Logger.new("vagrant::provisioners::bosh::deployment::manifest")
        end

        # Syncs releases to guest FS and rewrites manifest to reference guest FS locations.
        def resolve_releases
          uploaded_releases = uploadable_releases.map(&:upload)

          parsed_releases.each do |release|
            uploaded_releases.each do |uploaded_release|
              if release["name"] == uploaded_release.name
                release.merge!(uploaded_release.as_hash)
              end
            end
          end

          nil # ah
        end

        def as_string
          YAML.dump(parsed_manifest)
        end

        private

        # Returns releases with url matching `dir+bosh://...`
        def uploadable_releases
          parsed_releases.map { |release|
            next unless url = release["url"]
            next unless url =~ %r{\Adir\+bosh://(.+)\z}

            @uploadable_release_factory.new_uploadable_release(
              release["name"], 
              release["version"], 
              $1, # host_dir
            )
          }.compact
        end

        def parsed_releases
          parsed_manifest["releases"] || []
        end

        def parsed_manifest
          return @parsed_manifest if @parsed_manifest

          begin
            @parsed_manifest = YAML.load(@manifest)
          rescue SyntaxError => e
            error_msg = @ui.msg_string(:parse_error, details: e.inspect)
            raise VagrantPlugins::VagrantBosh::Errors::BoshReleaseError, error_msg
          end

          unless @parsed_manifest.is_a?(Hash)
            error_msg = @ui.msg_string(:non_hash_class_error, {
              actual_class: parsed_manifest.class.to_s,
            })
            raise VagrantPlugins::VagrantBosh::Errors::BoshReleaseError, error_msg
          end

          @parsed_manifest
        end
      end

      #~
    end
  end
end
