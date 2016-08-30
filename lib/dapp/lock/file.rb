module Dapp
  # Lock
  module Lock
    # File
    class File < Base
      class << self
        attr_writer :counter

        def counter
          @counter ||= 0
        end
      end # << self

      attr_reader :lock_path

      def initialize(lock_path, name)
        super(name)

        @lock_path = Pathname.new(lock_path).tap(&:mkpath)
      end

      protected

      def _do_lock(timeout, on_wait, readonly)
        @file = ::File.open(lock_path.join(name), ::File::RDWR | ::File::CREAT, 0o644)

        begin
          mode = (readonly ? ::File::LOCK_SH : ::File::LOCK_EX)
          _waiting(timeout, on_wait) { @file.flock(mode) } unless @file.flock(mode | ::File::LOCK_NB)
        rescue ::Timeout::Error
          raise Dapp::Lock::Error::Timeout, code: :timeout,
                                            data: { name: name, timeout: timeout }
        end

        self.class.counter += 1
      end

      def _do_unlock
        @file.close
        @file = nil
        self.class.counter -= 1
      end
    end # File
  end # Lock
end # Dapp
