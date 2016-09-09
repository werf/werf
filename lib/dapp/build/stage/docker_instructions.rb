module Dapp
  module Build
    module Stage
      # DockerInstructions
      class DockerInstructions < Base
        def initialize(application)
          @prev_stage = GALatestPatch.new(application, self)
          @application = application
        end

        def dependencies
          [change_options]
        end

        def prepare_image
          super
          change_options.each do |k, v|
            image.public_send("add_change_#{k}", v)
          end
        end

        def log_image_details
          super
          log_image_instructions
        end

        def log_image_instructions
          return if (instructions = image.prepare_instructions(image.send(:change_options))).empty?
          application.project.log_info application.project.t(code: 'image.instructions')
          application.project.with_log_indent { application.project.log_info instructions.join("\n") }
        end

        private

        def change_options
          @change_options ||= application.config._docker._change_options.to_h.delete_if { |_, val| val.nil? || (val.respond_to?(:empty?) && val.empty?) }
        end
      end # DockerInstructions
    end # Stage
  end # Build
end # Dapp
