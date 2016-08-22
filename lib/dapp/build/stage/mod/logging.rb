module Dapp
  module Build
    module Stage
      # Mod
      module Mod
        # Logging
        module Logging
          def log_image_build(&image_build)
            case
            when empty?                 then log_state(:empty)
            when image.tagged?          then log_state(:using_cache)
            when should_be_not_present? then log_state(:not_present)
            when application.dry_run?   then log_state(:build, styles: { status: :success })
            else                             log_image_build_process(&image_build)
            end
          ensure
            log_build
          end

          def log_build
            application.with_log_indent do
              application.log_info application.t(code: 'image.signature', data: { signature: image_name })
              log_image_details unless empty?
            end if application.log? && application.log_verbose? && !should_be_quiet?
          end

          def log_image_details
            if image.tagged?
              log_image_created_at
              log_image_size
            end
            log_image_commands unless ignore_log_commands?
          end

          def log_image_created_at
            application.log_info application.t(code: 'image.info.created_at',
                                               data: { value: Time.parse(image.created_at).localtime })
          end

          def log_image_size
            if from_image.tagged? && !prev_stage.nil?
              size = image.size.to_f - from_image.size.to_f
              code = 'image.info.difference'
            else
              size = image.size
              code = 'image.info.size'
            end
            application.log_info application.t(code: code, data: { value: size.to_f.round(2) })
          end

          def log_image_commands
            return if (bash_commands = image.send(:bash_commands)).empty?
            application.log_info application.t(code: 'image.commands')
            application.with_log_indent { application.log_info bash_commands.join("\n") }
          end

          def log_name
            application.t(code: name, context: name_context)
          end

          def log_state(state_code, styles: {})
            application.log_state(log_name, state: application.t(code: state_code, context: 'state'), **styles) unless should_be_quiet?
          end

          def log_image_build_process
            return yield if should_be_quiet?
            application.log_process(log_name, process: application.t(code: 'status.process.building'), short: should_not_be_detailed?) do
              yield
            end
          end

          def name_context
            :stage
          end

          def ignore_log_commands?
            false
          end

          def should_be_skipped?
            image.tagged? && !application.log_verbose? && application.cli_options[:introspect_stage].nil?
          end

          def should_not_be_detailed?
            image.send(:bash_commands).empty?
          end

          def should_be_introspected?
            application.cli_options[:introspect_stage] == name && !application.dry_run? && !application.is_artifact
          end

          def should_be_quiet?
            false
          end
        end
      end # Mod
    end # Stage
  end # Build
end # Dapp
