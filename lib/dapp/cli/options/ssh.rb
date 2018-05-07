module Dapp
  class CLI
    module Options
      module Ssh
        def self.extended(klass)
          klass.class_eval do
            option :ssh_key,
                  long: '--ssh-key SSH_KEY',
                  description: ['Enable only specified ssh keys ',
                                '(use system ssh-agent by default)'].join,
                  default: nil,
                  proc: ->(v) { composite_options(:ssh_key) << v }
          end
        end
      end
    end
  end
end
