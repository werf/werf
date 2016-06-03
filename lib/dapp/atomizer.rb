module Dapp
  # "Transaction" journal with rollback (mainly to protect cache fill with unbuildable configuration)
  # TODO Restore deleted files
  # TODO  write path into journal
  # TODO  write backup file to restore from
  # TODO  restore not committed delete paths
  class Atomizer
    def initialize(builder, file_path, lock_timeout: 10)
      @builder = builder
      @file_path = file_path
      @lock_timeout = lock_timeout
      @file = open

      builder.register_atomizer self
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

    protected

    attr_reader :file_path
    attr_reader :builder

    attr_reader :lock_timeout
    attr_reader :file

    def open
      file = File.open(file_path, File::RDWR | File::CREAT, 0644)

      file.sync = true

      Timeout.timeout(lock_timeout) do
        file.flock(File::LOCK_EX)
      end

      if (not_commited_paths = file.read.lines.map(&:strip))
        FileUtils.rm_rf not_commited_paths
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
