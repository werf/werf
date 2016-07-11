module Dapp
  class Application
    include CommonHelper
    include Dapp::Filelock

    attr_reader :conf
    attr_reader :opts
    attr_reader :last_stage
    attr_reader :show_only
    attr_reader :ignore_git_fetch

    def initialize(conf:, opts:, ignore_git_fetch: false)
      @conf = conf
      @opts = opts

      opts[:build_path] = opts[:build_dir] || home_path('build')
      opts[:build_cache_path] = opts[:build_cache_dir] || home_path('build_cache')

      @last_stage = Build::Stage::Source5.new(self)
      @show_only = !!opts[:show_only]
      @ignore_git_fetch = ignore_git_fetch
    end

    def build_and_fixate!
      last_stage.build!
      last_stage.fixate!
    end

    def push!(image_name)
      raise "Application isn't built yet!" unless last_stage.image.exist? or show_only

      tags.each do |tag_name|
        image_with_tag = [image_name, tag_name].join(':')
        show_only ? log(image_with_tag) : last_stage.image.pushing!(image_with_tag)
      end
    end

    def local_git_artifact
      local_git_artifact_list.first
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

    def tags
      tags = []
      tags += opts[:tag]
      tags << local_git_artifact.latest_commit if opts[:tag_commit]
      if opts[:tag_branch] and !(branch = local_git_artifact.repo.branch).nil?
        raise "Application has specific revision that isn't associated with a branch name!" if branch == 'HEAD'
        tags << branch
      end
      # tags << nil if opts[:tag_build_id] TODO
      # tags << nil if opts[:tag_ci] TODO
      tags << :latest if tags.empty?
      tags
    end

    def builder
      @builder ||= case conf.builder
        when :chef then Builder::Chef.new(self)
        else Builder::Shell.new(self)
      end
    end
  end # Application
end # Dapp
