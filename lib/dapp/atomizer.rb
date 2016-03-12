module Dapp
  # "Transaction" journal with rollback (mainly to protect cache fill with unbuildable configuration)
  class Atomizer
    class << self
      attr_accessor :lock_timeout
    end

    def initialize(builder, file_path)
      @builder = builder
      @file_path = file_path
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

    attr_reader :file

    def open
      file = File.open(file_path, File::RDWR | File::CREAT, 0644)

      file.sync = true

      Timeout.timeout(self.class.lock_timeout || 10) do
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
