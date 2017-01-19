module Dapp
  module Build
    module Stage
      # Mod
      module Mod
        # Logging
        module Logging
          def log_image_build(&image_build)
            if empty?                            then log_state(:empty)
            elsif image.built?                   then log_state(:using_cache)
            elsif should_be_not_present?         then log_state(:not_present)
            elsif dimg.project.dry_run?          then log_state(:build, styles: { status: :success })
            else log_image_build_process(&image_build)
            end
          ensure
            log_build
          end

          def log_build
            dimg.project.with_log_indent do
              dimg.project.log_info dimg.project.t(code: 'image.signature', data: { signature: image_name })
              log_image_details unless empty?
            end if dimg.project.log_verbose? && !should_be_quiet?
          end

          def log_image_details
            if image.tagged?
              log_image_created_at
              log_image_size
            end
            log_image_commands unless ignore_log_commands?
            log_image_instructions
          end

          def log_image_instructions
            return if (instructions = image.prepare_instructions(image.send(:change_options))).empty?
            dimg.project.log_info dimg.project.t(code: 'image.instructions')
            dimg.project.with_log_indent { dimg.project.log_info instructions.join("\n") }
          end

          def log_image_created_at
            dimg.project.log_info dimg.project.t(code: 'image.info.created_at',
                                                 data: { value: Time.parse(image.created_at).localtime })
          end

          def log_image_size
            if !prev_stage.nil? && from_image.tagged?
              bytes = image.size - from_image.size
              code = 'image.info.difference'
            else
              bytes = image.size
              code = 'image.info.mb_size'
            end
            dimg.project.log_info dimg.project.t(code: code, data: { mb: (bytes / 1000 / 1000).round(3) })
          end

          def log_image_commands
            return if (bash_commands = image.send(:bash_commands)).empty?
            dimg.project.log_info dimg.project.t(code: 'image.commands')
            dimg.project.with_log_indent { dimg.project.log_info bash_commands.join("\n") }
          end

          def log_name
            dimg.project.t(code: name, context: log_name_context)
          end

          def log_name_context
            :stage
          end

          def log_state(state_code, styles: {})
            dimg.project.log_state(log_name,
                                   state: dimg.project.t(code: state_code, context: 'state'),
                                   styles: styles) unless should_be_quiet?
          end

          def log_image_build_process
            return yield if should_be_quiet?
            dimg.project.log_process(log_name, process: dimg.project.t(code: 'status.process.building'),
                                               short: should_not_be_detailed?) do
              yield
            end
          end

          def ignore_log_commands?
            false
          end

          def should_not_be_detailed?
            image.send(:bash_commands).empty?
          end

          def should_be_introspected?
            dimg.project.cli_options[:introspect_stage] == name && !dimg.project.dry_run? && !dimg.artifact?
          end

          def should_be_quiet?
            dimg.artifact? && !dimg.project.log_verbose?
          end
        end
      end # Mod
    end # Stage
  end # Build
end # Dapp
