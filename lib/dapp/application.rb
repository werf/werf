module Dapp
  class Application
    include CommonHelper
    include Dapp::Filelock

    attr_reader :conf
    attr_reader :opts
    attr_reader :last_stage

    def initialize(conf:, opts:)
      @conf = conf
      @opts = opts

      opts[:log_indent] = 0

      opts[:build_path] = opts[:build_dir] || home_path('build')
      opts[:build_path] = build_path opts[:basename] if opts[:shared_build_dir]

      opts[:build_cache_path] = opts[:build_cache_path] || home_path('build_cache')

      @last_stage = Build::Stage::Source5.new(self)
    end

    def build_and_fixate!
      last_stage.build!
      last_stage.fixate!
    end

    def git_artifact_list
      [*local_git_artifact_list, *remote_git_artifact_list].compact
    end

    def local_git_artifact_list
      @local_git_artifact_list ||= Array(conf.git_artifact.local).map do |ga_conf|
        repo = GitRepo::Own.new(self)
        GitArtifact.new(repo, **ga_conf.artifact_options)
      end
    end

    def remote_git_artifact_list
      @remote_git_artifact_list ||= Array(conf.git_artifact.remote).map do |ga_conf|
        repo = GitRepo::Remote.new(self, ga_conf.name, url: ga_conf.url, ssh_key_path: ga_conf.ssh_key_path)
        repo.fetch!(ga_conf.branch)
        GitArtifact.new(repo, **ga_conf.artifact_options)
      end
    end

    def build_cache_path(*path)
      path.compact.map(&:to_s).inject(Pathname.new(opts[:build_cache_path]), &:+).expand_path.tap do |p|
        FileUtils.mkdir_p p.parent
      end
    end

    def home_path(*path)
      path.compact.map(&:to_s).inject(Pathname.new(conf.home_path), &:+).expand_path
    end

    def build_path(*path)
      path.compact.map(&:to_s).inject(Pathname.new(opts[:build_path]), &:+).expand_path.tap do |p|
        FileUtils.mkdir_p p.parent
      end
    end

    def container_build_path(*path)
      path.compact.map(&:to_s).inject(Pathname.new('/.build'), &:+)
    end

    def builder
      @builder ||= case conf.builder
        when :chef  then Builder::Chef.new(self)
        when :shell then Builder::Shell.new(self)
        else raise 'builder type is not defined!'
      end
    end
  end # Application
end # Dapp
