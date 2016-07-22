module Dapp
  # Application
  class Application
    # Logging
    module Logging
      def log?
        !log_quiet?
      end

      def log_quiet?
        cli_options[:log_quiet]
      end

      def log_verbose?
        cli_options[:log_verbose]
      end

      def log_time?
        cli_options[:log_time]
      end

      def dry_run?
        cli_options[:dry_run]
      end

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

        message = slice(message, state)
        state   = rjust(state, message)
        formatted_message = paint_string(message, styles[:message])
        formatted_status  = paint_string(state, styles[:status])

        log "#{formatted_message}#{formatted_status}"
      end

      def log_process(message, process: nil, short: false, style: {}, &blk)
        style[:message] ||= DEFAULT_STYLE[:message]
        style[:process] ||= DEFAULT_STYLE[:process]
        style[:failed] ||= DEFAULT_STYLE[:failed]
        style[:success] ||= DEFAULT_STYLE[:success]

        if log_verbose? && !short
          process ||= t('status.process.default')
          log_process_verbose(message, process: process, style: style, &blk)
        else
          log_process_short(message, style: style, &blk)
        end
      end

      def log_secondary_process(message, **kvargs, &blk)
        log_process(message, **kvargs.merge(style: { message: :secondary, success: :secondary }), &blk)
      end

      def log_process_verbose(message, process:, style: {}, &blk)
        success_status = t('status.success.default')
        failed_status = t('status.failed.default')
        message         = paint_string(message, style[:message])
        process         = paint_string(rjust(process, message), style[:process])
        info            = message + process
        success_message = slice(message, success_status) + paint_string(rjust(success_status, message), style[:success])
        failed_message  = paint_string(slice(message, failed_status) + rjust(failed_status, message), style[:failed])
        log_process_default(info, success_message, failed_message, &blk)
      end

      def log_process_short(message, style: {}, &blk)
        success_status = t('status.success.default')
        failed_status = t('status.failed.default')
        message         = paint_string(message, style[:message])
        longest_status  = (success_status.length >= failed_status.length) ? success_status : failed_status
        info            = "#{slice(message, longest_status)} ... "
        success_message = paint_string(rjust(success_status, info), style[:success])
        failed_message  = paint_string(rjust(failed_status, info), style[:failed])
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
        log "#{message} #{time}", indent: !inline, time: !inline
      end

      def rjust(str, start_string)
        str.rjust(free_space(start_string))
      end

      def slice(str, status)
        str.slice(0..free_space(status))
      end

      def free_space(str)
        base_time = log_time? ? log_time.length : 0
        indent = log_indent.length
        str = Paint.unpaint(str.to_s).length
        time = 15
        terminal_width - base_time - str - indent - time
      end

      def terminal_width
        @terminal_width ||= `tput cols`.strip.to_i
      end
    end # Logging
  end # Application
end # Dapp
