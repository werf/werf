module Dapp
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

    def build!
      begin
        run!
        @built_id = commit!
      ensure
        shellout("docker rm #{container_name}")
      end
    end

    def export!(name)
      image = self.class.new(built_id: built_id, name: name)
      image.tag!
      image.push!
      image.untag!
    end

    def tag!
      unless (existed_id = id).empty?
        raise 'Image with other id has already tagged' if built_id != existed_id
        return
      end
      shellout!("docker tag #{built_id} #{name}")
    end

    protected

    attr_reader :container_name
    attr_reader :bash_commands
    attr_reader :options

    def add_option(key, value)
      options[key] = (options[key].nil? ? value : (Array(options[key]) << value).flatten)
    end

    def run!
      raise '`from.built_id` is not defined!' if from.built_id.empty?
      shellout!("docker run #{prepared_options} --name=#{container_name} #{from.built_id} #{prepared_bash_command}")
    end

    def commit!
      shellout!("docker commit #{container_name}").stdout.strip
    end

    def prepared_options
      options.map { |k, vals| Array(vals).map{|v| "--#{k}=#{v}" }.join(' ') }.join(' ')
    end

    def prepared_bash_command
      "bash #{ "-lec \"eval $(echo #{Base64.strict_encode64(bash_commands.join('; '))} | base64 --decode)\"" unless bash_commands.empty? }"
    end
  end # StageImage
end # Dapp
