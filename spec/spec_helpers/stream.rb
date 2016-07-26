require 'stringio'

module SpecHelpers
  module Stream
    def capture_stdout(&blk)
      old = $stdout
      $stdout = fake = StringIO.new
      blk.call
      fake.string
    ensure
      $stdout = old
    end

    def capture_stderr(&blk)
      old = $stderr
      $stderr = fake = StringIO.new
      blk.call
      fake.string
    ensure
      $stderr = old
    end
  end
end
