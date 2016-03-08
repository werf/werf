module Dapper
  # Dockerfile builder and docker binary wrapper
  class Docker
    def initialize(builder)
      @builder = builder
    end

    def name(name = nil)
      @name = name || @name
    end

    def from(from = nil)
      @from = from || @from
    end

    def build_path(*path)
      builder.build_path(builder.home_branch, [name, 'docker'].compact.join('.'), *path)
    end

    def run(*command, step: :build)
      add_instruction step, :run, command.join(' && ')
    end

    def workdir(directory, step: :build)
      add_instruction step, :workdir, directory
    end

    def copy(what, where, step: :build)
      add_instruction step, :copy, what, where
    end

    def add(what, where, step: :build)
      add_instruction step, :add, what, where
    end

    def add_artifact(file_path, filename, where, step: :build)
      add_instruction step, :add_artifact, file_path, filename, where
    end

    def expose(*ports, step: :build)
      add_instruction step, :expose, ports
    end

    def env(step: :build, **env)
      add_instruction step, :env, env
    end

    def volume(*paths, step: :build)
      add_instruction step, :volume, paths
    end

    def cmd(*commands, step: :build)
      add_instruction step, :cmd, commands
    end

    def initialize_dup(other)
      super

      @name = @name.dup if @name
      @from = @from.dup if @from

      @instructions = @instructions.dup
      @instructions.each do |step, step_instructions|
        @instructions[step] = step_instructions.dup
      end
    end

    def build
      # prepare
      generate_dockerfile

      # build
      res = docker("build --force-rm=true --rm=true #{build_path}", log_verbose: true)

      # return image id
      res.stdout.lines.grep(/^Successfully built ([0-9a-f]+)\n$/).first.strip.split.last
    end

    def image_exists?(**kwargs)
      !!image_id(**kwargs)
    end

    def image_id(**kwargs)
      if (image = images(**kwargs).first)
        image[:id]
      end
    end

    def images(name:, tag: nil, registry: nil)
      docker('images').stdout.lines.drop(1).map(&:strip)
        .map { |line| Hash[[:name, :tag, :id].zip(line.strip.split(/\s{2,}/)[0..3])] }
        .select { |i| i[:name] == pad_image_name(name: name, registry: registry) && (!tag || i[:tag] == tag) }
    end

    def tag(origin, new, force: true)
      docker "tag#{" -f" if force} #{origin.is_a?(String) ? origin : pad_image_name(**origin)} #{pad_image_name(**new)}"
    end

    def rmi(**kwargs)
      docker "rmi #{pad_image_name(**kwargs)}"
    end

    def push(name:, tag: nil, registry: nil)
      docker "push #{pad_image_name name: name, tag: tag, registry: registry}", log_verbose: true
    end

    protected

    def pad_image_name(name:, tag: nil, registry: nil)
      name = "#{registry}/#{name}" if registry
      name = "#{name}:#{tag}" if tag
      name
    end

    attr_reader :builder

    def instructions(step)
      (@instructions ||= {})[step] ||= []
    end

    def add_instruction(step, *args)
      instructions(step) << args
    end

    def docker(command, **kwargs)
      builder.shellout "docker #{command}", **kwargs
    end

    def dockerfile_path
      build_path 'Dockerfile'
    end

    def generate_dockerfile
      File.open dockerfile_path, 'w' do |dockerfile|
        dockerfile.puts 'FROM ' + from

        [:begining, :prepare, :build, :setup].each do |step|
          instructions(step).each do |directive, *params|
            case directive
            when :run
              dockerfile.puts 'RUN ' + params[0]
            when :copy
              dockerfile.puts "COPY #{params[0]} #{params[1]}"
            when :add
              dockerfile.puts "ADD #{params[0]} #{params[1]}"
            when :add_artifact
              FileUtils.link params[0], build_path(params[1]), force: true
              dockerfile.puts "ADD #{params[1]} #{params[2]}"
            when :expose
              dockerfile.puts 'EXPOSE ' + params[0].map(&:to_s).join(' ')
            when :env
              dockerfile.puts 'ENV ' + params[0].map { |k, v| %(#{k}="#{v}") }.join(' ')
            when :volume
              dockerfile.puts 'VOLUME ' + params[0].join(' ')
            when :workdir
              dockerfile.puts "WORKDIR  #{params[0]}"
            when :cmd
              dockerfile.puts 'CMD ' + params[0].join(' ')
            end
          end
        end
      end
    end
  end
end
