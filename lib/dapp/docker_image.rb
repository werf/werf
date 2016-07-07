module Dapp
  class DockerImage
    include CommonHelper

    attr_reader :from
    attr_reader :container_name
    attr_reader :name
    attr_reader :bash_commands
    attr_reader :options

    def initialize(name:, from: nil)
      @from = from
      @bash_commands = []
      @options = {}
      @name = name
      @container_name = SecureRandom.hex
    end

    def built_id
      @built_id ||= id
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
      @bash_commands += commands.flatten
    end

    def exist?
      !id.empty?
    end

    def pull_and_set!
      pull!
      @built_id = id
    end

    def build!
      @built_id = if bash_commands.empty?
        from.built_id
      else
        begin
          run!
          commit!
        ensure
          shellout("docker rm #{container_name}")
        end
      end
    end

    def rmi!
      shellout!("docker rmi -f #{name}")
    end

    def tag!
      raise '`built_id` is not defined!' if built_id.empty?
      shellout!("docker tag #{built_id} #{name}")
    end

    def push!
      # TODO
    end

    protected

    def add_option(key, value)
      options[key] = (options[key].nil? ? value : (Array(options[key]) << value).flatten)
    end

    private

    def id
      shellout!("docker images -q --no-trunc=true #{name}").stdout.strip
    end

    def run!
      raise '`from.built_id` is not defined!' if from.built_id.empty?
      shellout!("docker run #{prepared_options} --name=#{container_name} #{from.built_id} #{prepared_bash_command}")
    end

    def pull!
      shellout!("docker pull #{name}")
    end

    def commit!
      shellout!("docker commit #{container_name}").stdout.strip
    end

    def prepared_options
      options.map { |k, vals| Array(vals).map{|v| "--#{k}=#{v}" }.join(' ') }.join(' ')
    end

    def prepared_bash_command
      "bash #{ "-lec \"#{prepared_commands}\"" unless bash_commands.empty? }"
    end

    def prepared_commands
      bash_commands.map { |command| command.gsub(/(\$|")/) { "\\#{$1}" } }.join('; ')
    end
  end # DockerImage
end # Dapp
