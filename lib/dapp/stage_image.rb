module Dapp
  # StageImage
  class StageImage < DockerImage
    def initialize(name:, built_id: nil, from: nil)
      @bash_commands = []
      @options = {}
      @container_name = SecureRandom.hex
      @built_id = built_id
      super(name: name, from: from)
    end

    def add_expose(value)
      add_option(:expose, value)
    end

    def add_volume(value)
      add_option(:volume, value)
    end

    def add_volumes_from(value)
      add_option(:'volumes-from', value)
    end

    def add_env(value)
      add_option(:env, value)
    end

    def add_commands(*commands)
      @bash_commands += commands.flatten
    end

    def built_id
      @built_id ||= id
    end

    def build!(log_verbose)
      run!(log_verbose)
      @built_id = commit!
    ensure
      shellout("docker rm #{container_name}")
    end

    def export!(name, log_verbose: false)
      image = self.class.new(built_id: built_id, name: name)
      image.tag!
      image.push!(log_verbose)
      image.untag!
    end

    def tag!(log_verbose = false)
      unless (existed_id = id).nil?
        ::Dapp::Application.error! 'Image with other id has already tagged' if built_id != existed_id
        return
      end
      shellout!("docker tag #{built_id} #{name}", log_verbose: log_verbose)
    end

    protected

    attr_reader :container_name
    attr_reader :bash_commands
    attr_reader :options

    def add_option(key, value)
      options[key] = (options[key].nil? ? value : (Array(options[key]) << value).flatten)
    end

    def run!(log_verbose = false)
      ::Dapp::Application.error! '`from.built_id` is not defined!' if from.built_id.nil?
      shellout!("docker run #{prepared_options} --name=#{container_name} #{from.built_id} #{prepared_bash_command}", log_verbose: log_verbose)
    end

    def commit!
      shellout!("docker commit #{container_name}").stdout.strip
    end

    def prepared_options
      options.map { |k, vals| Array(vals).map { |v| "--#{k}=#{v}" }.join(' ') }.join(' ')
    end

    def prepared_bash_command
      "bash #{"-lec \"eval $(echo #{Base64.strict_encode64(prepared_commands.join(' && '))} | base64 --decode)\"" unless bash_commands.empty?}"
    end

    def prepared_commands
      bash_commands.map { |command| command.gsub(/^[\ |;]*|[\ |;]*$/, '') } # strip [' ', ';']
    end
  end # StageImage
end # Dapp
