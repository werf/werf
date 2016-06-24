module Dapp
  class Docker
    attr_reader :atomizer

    def initialize(socket: nil, build: nil)
      @socket = socket || '/var/run/docker.sock'
      @atomizer = build.builder.register_docker_atomizer(build.build_path("#{build.signature}.docker_atomizer"))
    end

    # FIXME build!
    def build_image!(image_specification:, image_name:)
      container_name = SecureRandom.hex
      run(image_specification.options, container_name, image_specification.from, image_specification.bash_commands)
      # FIXME @id = ...
      commit(container_name, image_name)
      atomizer << image_name
    ensure
      rm(container_name)
    end

    def image_exist?(name)
      # FIXME shellout
      not image_info(name).nil?
    end

    protected

    def run(options, image_name, from_image_name, commands)
      # FIXME shellout
      Mixlib::ShellOut.new(['docker run',
                            prepare_options(options),
                            prepare_name(image_name),
                            from_image_name,
                            prepare_bash_command(commands)].join(' ')).run_command.tap(&:error!)
    end

    # FIXME commit!
    def commit(container_name, image_name)
      # FIXME shellout
      Mixlib::ShellOut.new(['docker commit', container_name, image_name].join(' ')).run_command.tap(&:error!)
    end

    def rm(container_name)
      # FIXME shellout
      Mixlib::ShellOut.new(['docker rm', container_name].join(' ')).run_command
    end

    def prepare_options(options)
      options.map{ |k, vals| Array(vals).map{|v| "--#{k}=#{v}" }.join(' ') }
    end

    def prepare_name(name)
      "--name=#{name}"
    end

    def prepare_bash_command(commands)
      "bash #{ "-lec \"#{prepare_commands(commands)}\"" unless commands.empty? }"
    end

    def prepare_commands(commands)
      commands.map { |command| command.gsub('$', '\$') }.join('; ')
    end

    def image_info(name)
      resp_if_success raw_connection.request(method: :get, path: "/images/#{name}/json")
    end

    def raw_connection
      Excon.new('unix:///', socket: @socket)
    end

    def resp_if_success(resp)
      JSON.load(resp.body) if resp.status == 200
    end
  end # Docker
end # Dapp
