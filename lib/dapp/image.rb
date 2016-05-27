module Dapp
  class Image
    include CommonHelper

    attr_reader   :from
    attr_reader   :build_cmd
    attr_accessor :build_opts

    def initialize(from:)
      @from = from
      @build_cmd = []
      @build_opts = {}
    end

    def build_cmd!(*cmd)
      build_cmd.push *cmd
    end

    def build_opts!(**options)
      build_opts.merge! **options
    end

    def signature
      sha256([from, build_cmd, build_opts.inspect])
    end
  end # Image
end # Dapp
