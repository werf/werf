module Dapp
  # Application
  class Application
    # Logging
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

      def dry_run
        cli_options[:dry_run]
      end

      DEFAULT_STYLE = {
        message: :step,
        process: :secondary,
        status:  :secondary,
        success: :success,
        failed:  :failed,
        time:    :default
      }.freeze

      DEFAULT_STATUS = {
        process: 'RUNNING',
        success: 'OK',
        failed: 'FAILED'
      }

      def log_state(message, state:, styles: {})
        styles[:message] ||= DEFAULT_STYLE[:message]
        styles[:status] ||= DEFAULT_STYLE[:status]

        state            = rjust("[#{state}]", message)
        formatted_message = paint_string(message, styles[:message])
        formatted_status  = paint_string(state, styles[:status])

        log "#{formatted_message}#{formatted_status}"
      end

      def log_process(message, process: nil, short: false, status: {}, style: {}, &blk)
        status[:process]  = "[#{process || DEFAULT_STATUS[:process]}]"
        status[:success]  = "[#{status[:success] || DEFAULT_STATUS[:success]}]"
        status[:failed]   = "[#{status[:failed] || DEFAULT_STATUS[:failed]}]"
        style[:message] ||= DEFAULT_STYLE[:message]
        style[:process] ||= DEFAULT_STYLE[:process]
        style[:failed] ||= DEFAULT_STYLE[:failed]
        style[:success] ||= DEFAULT_STYLE[:success]

        if log_verbose && !short
          log_process_verbose(message, statuse: status, style: style, &blk)
        else
          log_process_short(message, statuse: status, style: style, &blk)
        end
      end

      def log_secondary_proccess(message, **kvargs, &blk)
        log_process(message, **kvargs.merge(style: { message: :secondary, success: :secondary }), &blk)
      end

      def log_process_verbose(message, statuse: {}, style: {}, &blk)
        message         = paint_string(message, style[:message])
        process         = paint_string(rjust(statuse[:process], message), style[:process])
        info            = message + process
        success_message = message + paint_string(rjust(statuse[:success], message), style[:success])
        failed_message  = paint_string(message + rjust(statuse[:failed], message), style[:failed])
        log_process_default(info, success_message, failed_message, &blk)
      end

      def log_process_short(message, statuse: {}, style: {}, &blk)
        message         = paint_string(message, style[:message])
        info            = "#{message} ... "
        success_message = paint_string(rjust(statuse[:success], info), style[:success])
        failed_message  = paint_string(rjust(statuse[:failed], info), style[:failed])
        log_process_default(info, success_message, failed_message, inline: true, &blk)
      end

      def log_process_default(info, success_message, failed_message, inline: false)
        log info, inline: inline
        message = success_message
        start = Time.now
        yield
      rescue Error::Base, SignalException, StandardError => _e
        message = failed_message
        raise
      ensure
        time = paint_string("#{(Time.now - start).round(2)} sec", DEFAULT_STYLE[:time])
        log "#{message} #{time}", indent: !inline
      end

      def rjust(str, start_string)
        time = 20
        indent = log_indent.length
        start_string = Paint.unpaint(start_string.to_s).length
        str.rjust(terminal_width - start_string - indent - time)
      end

      def terminal_width
        @terminal_width ||= `tput cols`.strip.to_i
      end
    end # Logging
  end # Application
end # Dapp
