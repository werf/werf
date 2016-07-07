module SpecHelpers
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

    def self.included(base)
      base.extend(self)
    end
  end
end
