require "log4r"

module VagrantPlugins
  module VagrantBosh
    class Communicator
      def initialize(machine, ui)
        @c = machine.communicate

        @ui = ui.for(:communicator)
        @logger = Log4r::Logger.new("vagrant::provisioners::bosh::communicator")
      end

      def mkdir_p(path)
        debug_sudo("mkdir -p #{path}")
      end

      def rm_rf(path)
        debug_sudo("rm -rf #{path}")
      end

      def upload(src_path, dst_path)
        @ui.debug_msg(:upload, src_path: src_path, dst_path: dst_path)
        @c.upload(src_path, dst_path)
      end

      def mv(src_path, dst_path)
        debug_sudo("mv #{src_path} #{dst_path}")
      end

      def chmod_x(path)
        debug_sudo("chmod +x #{path}")
      end

      def chown(user, group, path, recursive=false)
        debug_sudo("chown #{"-R" if recursive} #{user}:#{group} #{path}")
      end

      def sudo(cmd, &blk)
        debug_sudo(cmd, &blk)
      end

      private

      def debug_sudo(cmd, &blk)
        key = cmd.include?("\n") ? :multi_line_sudo : :sudo
        @ui.debug_msg(key, cmd: cmd)        
        @c.sudo(cmd, nil, &blk)
      end
    end
  end
end
