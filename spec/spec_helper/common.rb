module SpecHelper
  module Common
    extend ActiveSupport::Concern

    included do
      before :each do
        @test_dir = Dir.pwd
      end
    end

    def shellout(*args, **kwargs)
      kwargs.delete :verbose
      Mixlib::ShellOut.new(*args, timeout: 20, **kwargs).run_command
    end

    def shellout!(*args, **kwargs)
      shellout(*args, **kwargs).tap(&:error!)
    end

    def host_docker
      ::Dapp::Dapp.host_docker
    end

    def random_string(n = 10)
      (('a'..'z').to_a * n).sample(n).join
    end

    def random_binary_string(n=10)
      ([0, 1] * n).sample(n).map(&:chr).join
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
