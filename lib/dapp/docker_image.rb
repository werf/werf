module Dapp
  class DockerImage
    include CommonHelper

    attr_reader :from
    attr_reader :container_name # FIXME private
    attr_reader :name
    attr_reader :bash_commands
    attr_reader :options
    attr_reader :application # FIXME remove

    def initialize(application, name:, built_id: nil, from: nil)
      @application = application

      @from = from
      @bash_commands = []
      @options = {}
      @name = name
      @built_id = built_id
      @container_name = SecureRandom.hex
    end

    def built_id
      @built_id || id
    end

    def id
      shellout!("docker images -q --no-trunc=true #{name}").stdout.strip
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

    def exist?
      !id.empty?
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

    def rmi! # FIXME ??? untag!
      return unless exist?
      shellout!("docker rmi #{name}")
    end

    def tag!
      unless (existed_id = id).empty?
        raise 'Image with other id has already tagged' if built_id != existed_id
        return
      end
      shellout!("docker tag #{built_id} #{name}")
    end

    def pull!
      shellout!("docker pull #{name}")
    end

    # FIXME image_name -> name
    def export!(image_name)
      image = self.class.new(built_id: built_id, name: image_name)
      image.tag!
      image.push!
      image.rmi!
    end

    def info
      raise "Image `#{name}` doesn't exist!" unless exist?
      date, bytesize = shellout!("docker inspect --format='{{.Created}} {{.Size}}' #{name}").stdout.strip.split
      ["date: #{Time.parse(date)}", "size: #{to_mb(bytesize.to_i)} MB"].join("\n")
    end

    protected

    def add_option(key, value)
      options[key] = (options[key].nil? ? value : (Array(options[key]) << value).flatten)
    end

    private

    def push!
      shellout!("docker push #{name}")
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
      "bash #{ "-lec #{prepared_script}" unless bash_commands.empty? }"
    end

    # FIXME use base64
    def prepared_script
      application.build_path("#{name}.sh").tap do |path|
        path.write <<BODY
#!bin/bash
#{bash_commands.join('; ')}
BODY
        path.chmod 0755
      end
      application.container_build_path("#{name}.sh")
    end
  end # DockerImage
end # Dapp
