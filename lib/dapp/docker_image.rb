module Dapp
  # DockerImage
  class DockerImage
    include Helper::Shellout

    attr_reader :from
    attr_reader :name

    def initialize(name:, from: nil)
      @from = from
      @name = name
    end

    def id
      @id || begin
        unless (output = shellout!("docker images -q --no-trunc=true #{name}").stdout.strip).empty?
          output
        end
      end
    end

    def untag!
      fail Error::Build, code: :image_is_already_untagged, data: { name: name } unless tagged?
      shellout!("docker rmi #{name}")
    end

    def push!(log_verbose: false, log_time: false)
      fail Error::Build, code: :image_is_not_exist, data: { name: name } unless tagged?
      shellout!("docker push #{name}", log_verbose: log_verbose, log_time: log_time)
    end

    def pull!(log_verbose: false, log_time: false)
      return if tagged?
      shellout!("docker pull #{name}", log_verbose: log_verbose, log_time: log_time)
      @pulled = true
    end

    def tagged?
      !!id
    end

    def pulled?
      !!@pulled
    end

    def info
      fail Error::Build, code: :image_is_not_exist, data: { name: name } unless tagged?
      shellout!("docker inspect --format='{{.Created}} {{.Size}}' #{name}").stdout.strip.split
    end
  end # DockerImage
end # Dapp
