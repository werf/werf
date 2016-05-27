module Dapp
  class Image
    attr_reader :from
    attr_reader :build_cmd

    def initialize(from:)
      @from = from
      @build_cmd = []
    end

    def build_cmd!(*cmd)
      build_cmd.push *cmd
    end
  end # Image
end # Dapp
