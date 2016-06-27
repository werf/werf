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

      opts[:build_path] = opts[:build_dir] ? opts[:build_dir] : home_path('build')
      opts[:build_path] = build_path opts[:basename] if opts[:shared_build_dir]

      @last_stage = Builder::Stage::Source5.new(self)
    end

    def build_and_fixate!
      last_stage.build!
      last_stage.fixate!
    end

    def git_artifact_list
      [local_git_artifact, *remote_git_artifact_list].compact
    end

    def local_git_artifact
      @local_git_artifact ||= begin
        cfg = (conf[:git_artifact] || {})[:local]
        repo = GitRepo::Own.new(self)
        GitArtifact.new(repo, cfg[:where_to_add], branch: cfg[:branch]) if cfg
      end
    end

    def remote_git_artifact_list
      @remote_git_artifact_list ||= Array((conf[:git_artifact] || {})[:remote]).map do |cfg|
        repo_name = cfg[:url].gsub(%r{.*?([^\/ ]+)\.git}, '\\1')
        repo = GitRepo::Remote.new(self, repo_name, url: cfg[:url], ssh_key_path: ssh_key_path)
        repo.fetch!(cfg[:branch])
        GitArtifact.new(repo, cfg[:where_to_add], branch: cfg[:branch])
      end
    end

    def home_path(*path)
      path.compact.inject(Pathname.new(conf[:home_path]), &:+).expand_path
    end

    def build_path(*path)
      path.compact.inject(Pathname.new(opts[:build_path]), &:+).expand_path.tap do |p|
        FileUtils.mkdir_p p.parent
      end
    end

    def container_build_path(*path)
      path.compact.inject(Pathname.new('/.build'), &:+).expand_path
    end

    def builder
      @builder ||= case conf[:type]
        when :chef then Builder::Chef.new(self)
        when :shell then Builder::Shell.new(self)
        else raise 'builder type is not defined!'
      end
    end
  end # Builder
end # Dapp
