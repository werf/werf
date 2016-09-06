module Dapp
  module Build
    module Stage
      # Mod
      module Mod
        # Logging
        module Logging
          def log_image_build(&image_build)
            if empty?                            then log_state(:empty)
            elsif image.tagged?                  then log_state(:using_cache)
            elsif should_be_not_present?         then log_state(:not_present)
            elsif application.project.dry_run?   then log_state(:build, styles: { status: :success })
            else log_image_build_process(&image_build)
            end
          ensure
            log_build
          end

          def log_build
            application.project.with_log_indent do
              application.project.log_info application.project.t(code: 'image.signature', data: { signature: image_name })
              log_image_details unless empty?
            end if application.project.log_verbose? && !should_be_quiet?
          end

          def log_image_details
            if image.tagged?
              log_image_created_at
              log_image_size
            end
            log_image_commands unless ignore_log_commands?
          end

          def log_image_created_at
            application.project.log_info application.project.t(code: 'image.info.created_at',
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
            application.project.log_info application.project.t(code: code, data: { value: size.to_f.round(2) })
          end

          def log_image_commands
            return if (bash_commands = image.send(:bash_commands)).empty?
            application.project.log_info application.project.t(code: 'image.commands')
            application.project.with_log_indent { application.project.log_info bash_commands.join("\n") }
          end

          def log_name
            application.project.t(code: name, context: log_name_context)
          end

          def log_name_context
            :stage
          end

          def log_state(state_code, styles: {})
            application.project.log_state(log_name,
                                          state: application.project.t(code: state_code, context: 'state'),
                                          styles: styles) unless should_be_quiet?
          end

          def log_image_build_process
            return yield if should_be_quiet?
            application.project.log_process(log_name, process: application.project.t(code: 'status.process.building'),
                                                      short: should_not_be_detailed?) do
              yield
            end
          end

          def ignore_log_commands?
            false
          end

          def should_be_skipped?
            image.tagged? && !application.project.log_verbose? && !should_be_introspected?
          end

          def should_not_be_detailed?
            image.send(:bash_commands).empty?
          end

          def should_be_introspected?
            application.project.cli_options[:introspect_stage] == name && !application.project.dry_run? && !application.is_artifact
          end

          def should_be_quiet?
            application.is_artifact && !application.project.log_verbose?
          end
        end
      end # Mod
    end # Stage
  end # Build
end # Dapp
