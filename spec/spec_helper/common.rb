module SpecHelper
  module Common
    def shellout(*args, **kwargs)
      kwargs.delete :log_verbose
      Mixlib::ShellOut.new(*args, timeout: 20, **kwargs).run_command
    end

    def shellout!(*args, **kwargs)
      shellout(*args, **kwargs).tap(&:error!)
    end

    def random_string
      (('a'..'z').to_a * 10).sample(100).join
    end

    def generate_command
      "echo '#{SecureRandom.hex}'"
    end

    def stub_instance(klass, &blk)
      method_new = klass.method(:new)
      stubbed_klass = class_double(klass).as_stubbed_const
      allow(stubbed_klass).to receive(:new) do |*args, &block|
        method_new.call(*args, &block).tap(&blk)
      end
    end

    def expect_exception_code(code)
      raise unless block_given?
      expect { yield }.to raise_error { |error| expect(error.net_status[:code]).to be(code) }
    end

    def self.included(base)
      base.extend(self)
    end
  end
end
