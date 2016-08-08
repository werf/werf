module Dapp
  module Build
    module Stage
      # Mod
      module Mod
        # Logging
        module Logging
          def log_build
            application.with_log_indent do
              application.log_info application.t(code: 'image.signature', data: { signature: image_name })
              unless image_empty?
                log_image_info
                log_image_commands
              end
            end if application.log? && application.log_verbose?
          end

          def log_image_commands
            return if (bash_commands = image.send(:bash_commands)).empty?
            application.log_info application.t(code: 'image.commands')
            application.with_log_indent { application.log_info bash_commands.join("\n") }
          end

          def log_image_info
            return unless image.tagged?
            date, size = image_info
            application.log_info application.t(code: 'image.info.date', data: { value: date })
            size_code = size_difference? ? 'image.info.difference' : 'image.info.size'
            application.log_info application.t(code: size_code, data: { value: size })
          end

          def image_info
            date, size = image.info
            if size_difference?
              _date, from_size = from_image.info
              size = size.to_f - from_size.to_f
            end

            [Time.parse(date).localtime, size.to_f.round(2)]
          end

          def size_difference?
            from_image.tagged? && !prev_stage.nil?
          end

          def should_be_skipped?
            image.tagged? && !application.log_verbose? && application.cli_options[:introspect_stage].nil?
          end

          def should_be_not_present?
            return false if next_stage.nil?
            next_stage.image.tagged? || next_stage.should_be_not_present?
          end

          def should_be_not_detailed?
            image.send(:bash_commands).empty?
          end

          def should_be_introspected?
            application.cli_options[:introspect_stage] == name && !application.dry_run? && !application.is_artifact
          end
        end
      end # Mod
    end # Stage
  end # Build
end # Dapp
