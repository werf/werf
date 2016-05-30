module Dapp
  class Docker
    def initialize(socket: nil)
      @socket = socket || '/var/run/docker.sock'
    end

    def raw_connection
      Excon.new('unix:///', socket: @socket)
    end

    def resp_if_success(resp)
      JSON.load(resp.body) if resp.status == 200
    end

    def build_image!(image:, name:)
      from = image.from
      cmd = image.build_cmd
      cmd.map! { |elm| elm.gsub('$', '\$') }
      docker_opts = image.build_opts
      container_name = SecureRandom.hex

      begin
        Mixlib::ShellOut.new(
          "docker run " +
          "#{docker_opts.map{|k, vals| Array(vals).map{|v| "--#{k}=#{v}"}}.join(' ')} " +
          "--name=#{container_name} " +
          "#{from} bash #{ "-lec \"#{cmd.join('; ')}\"" unless cmd.empty? }"
        ).run_command.tap(&:error!)
        Mixlib::ShellOut.new("docker commit #{container_name} #{name}").run_command.tap(&:error!)
      ensure
        Mixlib::ShellOut.new("docker rm #{container_name}").run_command
      end

      name
    end

    def image_exist?(name)
      not image_info(name).nil?
    end

    def image_info(name)
      resp_if_success raw_connection.request(method: :get, path: "/images/#{name}/json")
    end
  end # Docker
end # Dapp
