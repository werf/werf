module Dapp
  module Dimg
    module Build
      module Stage
        module Mod
          module Logging
            def log_image_build(&image_build)
              if empty?                    then log_state(:empty)
              elsif image.built?           then log_state(:using_cache)
              elsif should_be_not_present? then log_state(:not_present)
              elsif dimg.dapp.dry_run?     then log_state(:build, styles: { status: :success })
              else log_image_build_process(&image_build)
              end
            ensure
              log_build
            end

            def log_build
              return unless dimg.dapp.log_verbose? && !should_be_quiet?
              dimg.dapp.with_log_indent do
                dimg.dapp.log_info dimg.dapp.t(code: 'image.signature', data: { signature: image_name })
                log_image_details unless empty?
              end
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
              dimg.dapp.log_info dimg.dapp.t(code: 'image.instructions')
              dimg.dapp.with_log_indent { dimg.dapp.log_info instructions.join("\n") }
            end

            def log_image_created_at
              dimg.dapp.log_info dimg.dapp.t(code: 'image.info.created_at',
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
              dimg.dapp.log_info dimg.dapp.t(code: code, data: { mb: (bytes / 1000 / 1000).round(3) })
            end

            def log_image_commands
              return if (bash_commands = image.send(:bash_commands)).empty?
              dimg.dapp.log_info dimg.dapp.t(code: 'image.commands')
              dimg.dapp.with_log_indent { dimg.dapp.log_info bash_commands.join("\n") }
            end

            def log_name
              dimg.dapp.t(code: name, context: log_name_context)
            end

            def log_name_context
              :stage
            end

            def log_state(state_code, styles: {})
              return if should_be_quiet?
              dimg.dapp.log_state(log_name, state: dimg.dapp.t(code: state_code, context: 'state'), styles: styles)
            end

            def log_image_build_process
              return yield if should_be_quiet?
              dimg.dapp.log_process(log_name, process: dimg.dapp.t(code: 'status.process.building'), short: should_not_be_detailed?) { yield }
            end

            def ignore_log_commands?
              false
            end

            def should_not_be_detailed?
              image.send(:bash_commands).empty?
            end

            def should_be_introspected?
              dimg.stage_should_be_introspected?(name) && !dimg.dapp.dry_run?
            end

            def should_be_quiet?
              dimg.artifact? && !dimg.dapp.log_verbose?
            end
          end
        end # Mod
      end # Stage
    end # Build
  end # Dimg
end # Dapp
