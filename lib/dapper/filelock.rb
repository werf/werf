module Dapper
  module Filelock
    protected

    def filelocks
      @@filelocks ||= Hash.new(0)
    end

    def filelock(filelock, error_message: 'Already in use!', timeout: 10, &_block)
      File.open(build_path(filelock), File::RDWR | File::CREAT, 0644) do |file|
        Timeout.timeout(timeout) do
          file.flock(File::LOCK_EX) unless filelocks[filelock] > 0
        end

        begin
          filelocks[filelock] += 1
          yield
        ensure
          filelocks[filelock] -= 1
        end
      end
    rescue Timeout::Error
      STDERR.puts error_message
      exit 1
    end
  end
end
