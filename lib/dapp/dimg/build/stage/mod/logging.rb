module Dapp
  module Dimg
    module Build
      module Stage
        module Mod
          module Logging
            def log_image_build
              case
              when image.built?           then log_state(:using_cache)
              when should_be_not_present? then log_state(:not_present)
              when dimg.dapp.dry_run?     then log_state(:build, styles: { status: :success })
              else yield
              end
            ensure
              log_build
            end

            def log_build
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
                                             data: { value: Time.at(image.created_at).localtime })
            end

            def log_image_size
              if !prev_stage.nil? && from_image.tagged?
                bytes = image.size - from_image.size
                code = 'image.info.difference'
              else
                bytes = image.size
                code = 'image.info.mb_size'
              end
              dimg.dapp.log_info dimg.dapp.t(code: code, data: { mb: (bytes / 1000.0 / 1000.0).round(1) })
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
              dimg.dapp.log_state(log_name, state: dimg.dapp.t(code: state_code, context: 'state'), styles: styles)
            end

            def ignore_log_commands?
              false
            end

            def should_not_be_detailed?
              image.send(:bash_commands).empty?
            end

            def image_should_be_introspected?
              dimg.stage_should_be_introspected?(name) && !dimg.dapp.dry_run?
            end
          end
        end # Mod
      end # Stage
    end # Build
  end # Dimg
end # Dapp
