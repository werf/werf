module Dapp
  # Application
  class Application
    module Logging
      def log?
        !log_quiet
      end

      def log_quiet
        cli_options[:log_quiet]
      end

      def log_verbose
        cli_options[:log_verbose]
      end

      def log_state(message, status, indent: false, styles: {})
        styles[:message]  = style(styles[:message] || :step)
        styles[:status]   = style(styles[:status]  || :secondary)

        status            = rjust("[#{status}]", message)
        formatted_message = Paint[message.to_s, *styles[:message]]
        formatted_status  = Paint[status, *styles[:status]]

        log "#{formatted_message} #{formatted_status}", indent: indent
      end

      def log_process(message, process: 'RUNNING', indent: false, statuses: {}, styles: {})
        styles[:message] = style(styles[:message] || :step)
        styles[:process] = style(styles[:process] || :secondary)
        styles[:success] = style(styles[:success] || :success)
        styles[:failed]  = style(styles[:failed]  || :failed)
        styles[:time]    = styles[:time] ? style(styles[:time]) : :white

        message           = "#{message} ... " unless log_verbose
        status            = rjust("[#{statuses[:success] || 'OK'}]", message)
        process           = rjust("[#{process}]", message)
        formatted_message = Paint[message, *styles[:message]]
        formatted_process = Paint[process, *styles[:process]]
        formatted_status  = Paint[status, *styles[:success]]

        if log_verbose
          log "#{formatted_message} #{formatted_process}", indent: indent
        else
          log "#{formatted_message} ", new_line: false, indent: indent
        end

        start = Time.now
        yield
      rescue Exception => _e
        status  = rjust("[#{statuses[:failed] || 'FAILED'}]", message)
        formatted_status = Paint[status, *styles[:failed]]
        raise
      ensure
        time = Paint["#{(Time.now - start).round(2)} sec", styles[:time]]

        if log_verbose
          log "#{formatted_message} #{formatted_status} #{time}", indent: indent
        else
          log "#{formatted_status} #{time}", ignore_indent: true
        end
      end

      def log_secondary_proccess(message, **kvargs, &blk)
        styles = { styles: { message: :secondary, success: :secondary } }
        log_process(message, **kvargs.merge(styles), &blk)
      end

      def rjust(string, start_string)
        time = 20
        indent = log_indent.length
        start_string = start_string.length
        string.rjust(terminal_width - start_string - indent - time)
      end

      def terminal_width
        @terminal_width ||= `tput cols`.strip.to_i
      end
    end # Logging
  end # Application
end # Dapp
