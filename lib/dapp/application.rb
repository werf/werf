module Dapp
  # Application
  class Application
    include Helper::Log
    include Helper::Shellout
    include Helper::Sha256
    include Logging
    include GitArtifact
    include Path
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

    def build!
      last_stage.build!
      last_stage.save_in_cache!
    end

    def export!(repo)
      raise Error::Application, code: :application_is_not_built unless last_stage.image.tagged? || dry_run

      tags.each do |tag|
        image_name = [repo, tag].join(':')
        if dry_run
          log_state(image_name, 'PUSH', styles: { status: :success })
        else
          log_process(image_name, process: 'PUSHING') do
            last_stage.image.export!(image_name, log_verbose: log_verbose)
          end
        end
      end
    end

    def builder
      @builder ||= Builder.const_get(config._builder.capitalize).new(self)
    end

    def dry_run
      cli_options[:dry_run]
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
      raise Error::Application, code: :git_branch_without_name if (branch = git_repo.branch) == 'HEAD'
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
        raise Error::Application, code: :ci_environment_required
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
        raise Error::Application, code: :ci_environment_required
      end

      [branch, tag].compact
    end
  end # Application
end # Dapp
