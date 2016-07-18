module Dapp
  # Application
  class Application
    include CommonHelper
    include Dapp::Filelock

    attr_reader :config
    attr_reader :cli_options
    attr_reader :ignore_git_fetch

    def initialize(config:, cli_options:, ignore_git_fetch: false)
      @config = config
      @cli_options = cli_options

      @build_path = cli_options[:build_dir] || home_path('build')
      @build_cache_path = cli_options[:build_cache_dir] || home_path('build_cache')

      @last_stage = Build::Stage::Source5.new(self)
      @ignore_git_fetch = ignore_git_fetch
    end

    def dry_run
      !!cli_options[:dry_run]
    end

    def logging?
      cli_options[:log_verbose] || !cli_options[:log_quiet]
    end

    def build!
      last_stage.build!
      last_stage.save_in_cache!
    end

    def export!(repo)
      raise "Application isn't built yet!" unless last_stage.image.tagged? || dry_run

      tags.each do |tag|
        image_name = [repo, tag].join(':')
        dry_run ? log(image_name) : last_stage.image.export!(image_name)
      end
    end

    def git_artifacts
      [*local_git_artifacts, *remote_git_artifacts].compact
    end

    def local_git_artifacts
      @local_git_artifact_list ||= Array(config._git_artifact._local).map do |ga_config|
        repo = GitRepo::Own.new(self)
        GitArtifact.new(repo, **ga_config._artifact_options)
      end
    end

    def remote_git_artifacts
      @remote_git_artifact_list ||= Array(config._git_artifact._remote).map do |ga_config|
        repo = GitRepo::Remote.new(self, ga_config._name, url: ga_config._url, ssh_key_path: ga_config._ssh_key_path)
        repo.fetch!(ga_config._branch)
        GitArtifact.new(repo, **ga_config._artifact_options)
      end
    end

    def build_cache_path(*path)
      path.compact.map(&:to_s).inject(Pathname.new(@build_cache_path), &:+).expand_path.tap do |p|
        FileUtils.mkdir_p p.parent
      end
    end

    def home_path(*path)
      path.compact.map(&:to_s).inject(Pathname.new(config._home_path), &:+).expand_path
    end

    def build_path(*path)
      path
        .compact
        .map(&:to_s)
        .inject(Pathname.new(@build_path), &:+)
        .expand_path.tap { |p| FileUtils.mkdir_p p.parent }
    end

    def container_build_path(*path)
      path.compact.map(&:to_s).inject(Pathname.new('/.build'), &:+)
    end

    def builder
      @builder ||= Builder.const_get(config._builder.capitalize).new(self)
    end

    protected

    attr_reader :last_stage

    def git_repo
      @git_repo ||= GitRepo::Own.new(self)
    end

    def tags
      tags = simple_tags + branch_tags + commit_tags + build_tags + ci_tags
      tags << :latest if tags.empty?
      tags
    end

    def simple_tags
      cli_options[:tag]
    end

    def branch_tags
      return [] unless cli_options[:tag_branch]
      raise "Application has specific revision that isn't associated with a branch name!" if (branch = git_repo.branch) == 'HEAD'
      [branch]
    end

    def commit_tags
      return [] unless cli_options[:tag_commit]
      commit = git_repo.latest_commit
      [commit]
    end

    def build_tags
      return [] unless cli_options[:tag_build_id]

      if ENV['GITLAB_CI']
        build_id = ENV['CI_BUILD_ID']
      elsif ENV['TRAVIS']
        build_id = ENV['TRAVIS_BUILD_NUMBER']
      else
        raise 'CI environment required (Travis or GitLab CI)'
      end

      [build_id]
    end

    def ci_tags
      return [] unless cli_options[:tag_ci]

      if ENV['GITLAB_CI']
        branch = ENV['CI_BUILD_REF_NAME']
        tag = ENV['CI_BUILD_TAG']
      elsif ENV['TRAVIS']
        branch = ENV['TRAVIS_BRANCH']
        tag = ENV['TRAVIS_TAG']
      else
        raise 'CI environment required (Travis or GitLab CI)'
      end

      [branch, tag].compact
    end
  end # Application
end # Dapp
