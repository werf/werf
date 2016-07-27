require 'stringio'

module SpecHelpers
  module Stream
    def capture_stdout
      old = $stdout
      $stdout = fake = StringIO.new
      yield
      fake.string
    ensure
      $stdout = old
    end

    def capture_stderr
      old = $stderr
      $stderr = fake = StringIO.new
      yield
      fake.string
    ensure
      $stderr = old
    end
  end
end
