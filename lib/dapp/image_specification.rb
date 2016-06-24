module Dapp
  class ImageSpecification
    include CommonHelper

    attr_reader :from_name
    attr_reader :bash_commands
    attr_reader :options
    attr_reader :docker

    def initialize(from_name:)
      @from_name = from_name
      @bash_commands = []
      @options = {}
    end

    def add_expose(value)
      add_option(:expose, value)
    end

    def add_volume(value)
      add_option(:volume, value)
    end

    def add_env(value)
      add_option(:env, value)
    end

    def add_commands(*commands)
      bash_commands.push *commands
    end

    def signature
      hashsum [from_name, *bash_commands, options.inspect]
    end

    private

    def add_option(key, value)
      options[key] = (options[key].nil? ? value : (Array(options[key]) << value).flatten)
    end
  end # Image
end # Dapp
