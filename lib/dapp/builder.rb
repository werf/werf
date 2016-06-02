module Dapp
  module Builder
    def self.new(docker:, conf:, opts:)
      if conf[:type] == :chef
        Chef.new(docker: docker, conf: conf, opts: opts)
      elsif conf[:type] == :shell
        Shell.new(docker: docker, conf: conf, opts: opts)
      end
    end
  end # Builder
end # Dapp
