module Dapp
  module CommonHelper
    def log(message)
      return unless defined? opts
      puts ' ' * opts[:log_indent] + ' * ' + message if opts[:log_verbose] || !opts[:log_quiet]
    end

    def shellout(*args, log_verbose: false, **kwargs)
      log_verbose = (log_verbose and opts[:log_verbose]) if defined? opts
      kwargs[:live_stream] = STDOUT if log_verbose
      Mixlib::ShellOut.new(*args, timeout: 3600, **kwargs).run_command
    end

    def shellout!(*args, **kwargs)
      shellout(*args, **kwargs).tap(&:error!)
    end

    def hashsum(arg)
      sha256(arg)
    end

    def sha256(arg)
      Digest::SHA256.hexdigest Array(arg).compact.map(&:to_s).join(':::')
    end

    def kwargs(args)
      args.last.is_a?(Hash) ? args.pop : {}
    end

    def delete_file(path)
      path = Pathname(path)
      path.delete if path.exist?
    end
  end # CommonHelper
end # Dapp
