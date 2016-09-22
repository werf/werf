module Dapp
  # Project
  class Project
    # Logging
    module Logging
      # Process
      module Process
        DEFAULT_STYLE = {
          message: :step,
          process: :secondary,
          status:  :secondary,
          success: :success,
          failed:  :warning,
          time:    :default
        }.freeze

        def log_state(message, state:, styles: {})
          styles[:message] ||= DEFAULT_STYLE[:message]
          styles[:status] ||= DEFAULT_STYLE[:status]

          message = slice(message)
          state   = rjust(state, message)
          formatted_message = paint_string(message, styles[:message])
          formatted_status  = paint_string(state, styles[:status])

          log "#{formatted_message}#{formatted_status}"
        end

        # rubocop:disable Metrics/ParameterLists
        def log_process(message, process: nil, short: false, quiet: false, style: {}, status: {}, &blk)
          return yield if quiet

          style[:message] ||= DEFAULT_STYLE[:message]
          style[:process] ||= DEFAULT_STYLE[:process]
          style[:failed] ||= DEFAULT_STYLE[:failed]
          style[:success] ||= DEFAULT_STYLE[:success]

          status[:success] ||= t(code: 'status.success.default')
          status[:failed] ||= t(code: 'status.failed.default')

          if log_verbose? && !short
            process ||= t(code: 'status.process.default')
            log_process_verbose(message, process: process, style: style, status: status, &blk)
          else
            log_process_short(message, style: style, status: status, &blk)
          end
        end
        # rubocop:enable Metrics/ParameterLists

        def log_secondary_process(message, **kwargs, &blk)
          log_process(message, **kwargs.merge(style: { message: :secondary, success: :secondary }), &blk)
        end

        protected

        def log_process_verbose(message, process:, style: {}, status: {}, &blk)
          process         = paint_string(rjust(process, message), style[:process])
          info            = paint_string(message, style[:message]) + process
          success_message = paint_string(slice(message), style[:message]) +
                            paint_string(rjust(status[:success], message), style[:success])
          failed_message  = paint_string(slice(message) + rjust(status[:failed], message), style[:failed])
          log_process_default(info, success_message, failed_message, &blk)
        end

        def log_process_short(message, style: {}, status: {}, &blk)
          info            = "#{paint_string(slice(message), style[:message])} ... "
          success_message = paint_string(rjust(status[:success], info), style[:success])
          failed_message  = paint_string(rjust(status[:failed], info), style[:failed])
          log_process_default(info, success_message, failed_message, inline: true, &blk)
        end

        def log_process_default(info, success_message, failed_message, inline: false)
          log info, inline: inline
          message = success_message
          start = Time.now
          yield
        rescue Exception::Base, Error::Base, SignalException, StandardError => _e
          message = failed_message
          raise
        ensure
          time = paint_string("#{(Time.now - start).round(2)} sec", DEFAULT_STYLE[:time])
          log "#{message} #{time}", indent: !inline, time: !inline
        end

        def rjust(str, start_string)
          str.rjust(free_space(start_string))
        end

        def slice(str)
          if (index = free_space(t(code: 'state.using_cache'))) >= 0 # free space by longest status
            str.slice(0..index)
          else
            str.slice(0)
          end
        end

        def free_space(str)
          base_time = log_time? ? log_time.length : 0
          indent = log_indent.length
          str = unpaint(str.to_s).length
          time = 15
          terminal_width - base_time - str - indent - time
        end

        def terminal_width
          @terminal_width ||= `tput cols`.strip.to_i
        end
      end
    end # Logging
  end # Project
end # Dapp
