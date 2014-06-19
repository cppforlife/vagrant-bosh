require "log4r"
require "i18n"

module VagrantPlugins
  module VagrantBosh
    class Ui
      def self.for_machine(machine)
        new(machine, {
          show_debug:  !!ENV["DEBUG"],
          i18n_prefix: nil,
          start_time:  Time.now,
        })
      end

      def initialize(machine, opts)
        @machine = machine

        @show_debug  = opts.fetch(:show_debug)
        @i18n_prefix = opts.fetch(:i18n_prefix)
        @start_time  = opts.fetch(:start_time)

        @logger = Log4r::Logger.new("vagrant::provisioners::bosh::ui")
      end

      def for(extra_i18n_prefix)
        self.class.new(@machine, { 
          show_debug:  @show_debug, 
          i18n_prefix: [@i18n_prefix, extra_i18n_prefix].compact.join("."),
          start_time:  @start_time,
        })
      end

      def section(key, hash={})
        msg(key, hash)
      end

      def msg(key, hash)
        path = @i18n_prefix ? "#{@i18n_prefix}.#{key}" : key
        title = I18n.t("bosh.ui.#{path}", hash)
        
        @machine.ui.info(time_prefix(title))
      end

      def timed_msg(key, hash, &blk)
        path = @i18n_prefix ? "#{@i18n_prefix}.#{key}" : key
        title = I18n.t("bosh.ui.#{path}", hash)

        # Do not show elapsed time in debug mode
        if @show_debug
          @machine.ui.info(time_prefix(title))
          blk.call
          return
        end

        # In non-debug mode show "Uploading /var/vcap/bosh/bin/bosh-agent (sudo)... 1.33s"
        begin
          t1 = Time.now
          @machine.ui.info(time_prefix("#{title}..."), new_line: false)
          blk.call
        ensure
          t2 = Time.now
          @machine.ui.info(" %.2fs" % (t2 - t1), prefix: false)
        end
      end

      def debug_msg(key, hash)
        msg(key, hash) if @show_debug
      end

      private

      def time_prefix(str)
        "[#{"%.2fs" % (Time.now - @start_time)}] #{str}"
      end
    end
  end
end
