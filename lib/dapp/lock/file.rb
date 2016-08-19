module Dapp
  # Lock
  module Lock
    # File
    class File < Base
      attr_reader :lock_path

      def initialize(lock_path, name, **kwargs)
        super(name, **kwargs)
        @lock_path = Pathname.new(lock_path).tap(&:mkpath)
      end

      def lock(shared: false)
        return if @file
        @file = ::File.open(lock_path.join(name), ::File::RDWR | ::File::CREAT, 0644)

        begin
          mode = (shared ? ::File::LOCK_SH : ::File::LOCK_EX)
          _waiting { @file.flock(mode) } unless @file.flock(mode | ::File::LOCK_NB)
        rescue ::Timeout::Error
          raise Dapp::Lock::Error::Timeout, code: :timeout,
                                            data: { name: name, timeout: timeout }
        end
      end

      def unlock
        @file.close
        @file = nil
      end
    end # File
  end # Lock
end # Dapp
