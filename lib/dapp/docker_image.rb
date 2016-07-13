module Dapp
  class DockerImage
    include CommonHelper

    attr_reader :from
    attr_reader :name

    def initialize(name:, from: nil)
      @from = from
      @name = name
    end

    def id
      @id ||= shellout!("docker images -q --no-trunc=true #{name}").stdout.strip
    end

    def untag!
      raise "Image `#{name}` is already untagged!" unless tagged?
      shellout!("docker rmi #{name}")
    end

    def push!
      raise "Image `#{name}` is not exist!" unless tagged?
      shellout!("docker push #{name}")
    end

    def pull!
      return if tagged?
      shellout!("docker pull #{name}")
      @pulled = true
    end

    def tagged?
      !id.empty?
    end

    def pulled?
      !!@pulled
    end

    def info
      raise "Image `#{name}` doesn't exist!" unless tagged?
      date, bytesize = shellout!("docker inspect --format='{{.Created}} {{.Size}}' #{name}").stdout.strip.split
      ["date: #{Time.parse(date)}", "size: #{to_mb(bytesize.to_i)} MB"].join("\n")
    end
  end # DockerImage
end # Dapp
