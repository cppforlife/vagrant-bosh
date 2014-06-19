require "log4r"
require "tempfile"
require "securerandom"

module VagrantPlugins
  module VagrantBosh
    class AssetUploader
      def initialize(communicator, ui, assets_path)
        @c = communicator
        @ui = ui.for(:asset_uploader)
        @assets_path = assets_path
        @logger = Log4r::Logger.new("vagrant::provisioners::bosh::asset_uploader")
      end

      def sync(dst_path)
        @ui.timed_msg(:upload, dst_path: dst_path) do
          upload_path(@assets_path, dst_path)
        end
      end

      def upload_text(text, dst_path)
        @ui.timed_msg(:upload, dst_path: dst_path) do
          begin
            f = Tempfile.new("asset-uploader-upload-text")
            f.write(text)
            f.flush
            upload_path(f.path, dst_path)
          ensure
            f.close if f
          end
        end
      end

      private

      def upload_path(src_path, dst_path)
        dst_tmp_path = "/tmp/#{SecureRandom.hex(5)}"

        @c.upload(src_path, dst_tmp_path)

        if File.directory?(src_path)
          @c.mkdir_p(dst_path) # create nested dst path
          @c.rm_rf(dst_path)
          @c.mv(dst_tmp_path, dst_path)
          @c.chown("root", "root", dst_path, true)
        else
          @c.mv(dst_tmp_path, dst_path)
          @c.chown("root", "root", dst_path)
        end
      end
    end
  end
end
