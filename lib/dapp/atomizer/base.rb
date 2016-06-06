module Dapp
  module Atomizer
    # "Transaction" journal with rollback (mainly to protect cache fill with unbuildable configuration)
    # TODO Restore deleted files
    # TODO  write path into journal
    # TODO  write backup file to restore from
    # TODO  restore not committed delete paths

    class Base
      include Dapp::CommonHelper

      def initialize(file_path, lock_timeout: 10)
        @file_path = file_path
        @lock_timeout = lock_timeout
        @file = open
      end

      def <<(path)
        file.puts path
        file.fsync
      end

      def commit!
        @file.truncate(0)
        @file.close
        @file = open
      end

      def rollback(lines)
        raise
      end

      protected

      attr_reader :file_path

      attr_reader :lock_timeout
      attr_reader :file

      def open
        file = ::File.open(file_path, ::File::RDWR | ::File::CREAT, 0644)

        file.sync = true

        Timeout.timeout(lock_timeout) do
          file.flock(::File::LOCK_EX)
        end

        if (lines = file.read.lines.map(&:strip))
          rollback(lines)
        end

        file.truncate(0)
        file.rewind

        file
      rescue Timeout::Error
        file.close

        STDERR.puts 'Atomizer already in use! Try again later.'
        exit 1
      end
    end
  end
end