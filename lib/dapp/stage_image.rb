module Dapp
  # StageImage
  class StageImage < DockerImage
    def initialize(name:, built_id: nil, from: nil)
      @bash_commands = []
      @options = {}
      @change_options = {}
      @container_name = SecureRandom.hex
      @built_id = built_id
      super(name: name, from: from)
    end

    def add_change_volume(value)
      add_change_option(:volume, value)
    end

    def add_change_expose(value)
      add_change_option(:expose, value)
    end

    def add_change_env(value)
      add_change_option(:env, value)
    end

    def add_change_label(value)
      add_change_option(:label, value)
    end

    def add_change_cmd(value)
      add_change_option(:cmd, value)
    end

    def add_change_onbuild(value)
      add_change_option(:onbuild, value)
    end

    def add_change_workdir(value)
      add_change_option(:workdir, value)
    end

    def add_change_user(value)
      add_change_option(:user, value)
    end

    def add_volume(value)
      add_option(:volume, value)
    end

    def add_volumes_from(value)
      add_option(:'volumes-from', value)
    end

    def add_commands(*commands)
      @bash_commands.concat(commands.flatten)
    end

    def built_id
      @built_id ||= id
    end

    def build!(**kvargs)
      @built_id = if should_be_built?
                    begin
                      run!(**kvargs)
                      commit!
                    ensure
                      shellout("docker rm #{container_name}")
                    end
                  else
                    from.built_id
                  end
    end

    def export!(name, log_verbose: false, log_time: false, force: false)
      image = self.class.new(built_id: built_id, name: name)
      image.tag!(log_verbose: log_verbose, log_time: log_time, force: force)
      image.push!(log_verbose: log_verbose, log_time: log_time)
      image.untag!
    end

    def tag!(log_verbose: false, log_time: false, force: false)
      if !(existed_id = id).nil? && !force
        raise Error::Build, code: :another_image_already_tagged if built_id != existed_id
        return
      end
      shellout!("docker tag #{built_id} #{name}", log_verbose: log_verbose, log_time: log_time)
    end

    protected

    attr_reader :container_name
    attr_reader :bash_commands
    attr_reader :options, :change_options

    def add_option(key, value)
      add_option_default(options, key, value)
    end

    def add_change_option(key, value)
      add_option_default(change_options, key, value)
    end

    def add_option_default(hash, key, value)
      hash[key] = (hash[key].nil? ? value : (Array(hash[key]) << value).flatten)
    end

    def run!(log_verbose: false, log_time: false, introspect_error: false, introspect_before_error: false)
      raise Error::Build, code: :built_id_not_defined if from.built_id.nil?
      shellout!("docker run #{prepared_options} --name=#{container_name} #{from.built_id} #{prepared_bash_command}",
                log_verbose: log_verbose, log_time: log_time)
    rescue Error::Shellout => e
      raise unless introspect_error || introspect_before_error
      built_id = introspect_error ? commit! : from.built_id
      raise Exception::IntrospectImage, message: Dapp::Helper::NetStatus.message(e),
                                        data: { built_id: built_id, options: prepared_options, rmi: introspect_error }
    end

    def commit!
      shellout!("docker commit #{prepared_change} #{container_name}").stdout.strip
    end

    def should_be_built?
      !(bash_commands.empty? && change_options.empty?)
    end

    def prepared_options
      prepared_options_default(options) { |k, vals| Array(vals).map { |v| "--#{k}=#{v}" }.join(' ') }
    end

    def prepared_change
      prepared_options_default(change_options) do |k, vals|
        if k == :cmd
          %(-c '#{k.to_s.upcase} #{Array(vals)}')
        else
          Array(vals).map { |v| %(-c "#{k.to_s.upcase} #{v}") }.join(' ')
        end
      end
    end

    def prepared_options_default(hash)
      hash.map { |k, vals| yield(k, vals) }.join(' ')
    end

    def prepared_bash_command
      "bash -ec 'eval $(echo #{Base64.strict_encode64(prepared_commands.join(' && '))} | base64 --decode)'"
    end

    def prepared_commands
      bash_commands.map { |command| command.gsub(/^[\ |;]*|[\ |;]*$/, '') } # strip [' ', ';']
    end
  end # StageImage
end # Dapp
