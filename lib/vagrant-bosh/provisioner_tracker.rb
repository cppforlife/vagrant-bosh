require "log4r"
require "json"

module VagrantPlugins
  module VagrantBosh
    class ProvisionerTracker
      def initialize(ui)
        @ui = ui.for(:provisioner_tracker)
        @logger = Log4r::Logger.new("vagrant::provisioners::bosh::provisioner_tracker")
      end

      def add_data(type, data)
        if type == :stdout
          data.split("\n").each { |raw_event| add_event(raw_event) }
        else
          add_debug(data)
        end
      end

      private

      def add_event(raw_event)
        begin
          event = JSON.parse(raw_event)
        rescue JSON::ParserError
          @ui.msg(:invalid_event, content: raw_event)
        else
          @ui.msg(:event, {
            state: event["state"].capitalize, # Started, Finished
            stage: event["stage"].downcase, 
            task:  event["task"],
          })
        end
      end

      def add_debug(data)
        @ui.debug_msg(:debug, content: data)
      end
    end
  end
end
