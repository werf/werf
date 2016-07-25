module Dapp
  module Build
    module Stage
      # Source5
      class Source5 < SourceBase
        def initialize(application)
          @prev_stage = Source4.new(application, self)
          @application = application
        end

        def prev_source_stage
          prev_stage
        end

        def next_source_stage
          nil
        end

        def signature
          hashsum [super, *exposes]
        end

        def image
          super do |image|
            image.add_expose(exposes)  unless exposes.empty?
            image.add_env(envs)        unless envs.empty?
            image.add_workdir(workdir) unless workdir.nil?
          end
        end

        protected

        def exposes
          application.config._docker._expose
        end

        def envs
          application.config._docker._env
        end

        def workdir
          application.config._docker._workdir
        end

        def layers_commits_write!
          nil
        end
      end # Source5
    end # Stage
  end # Build
end # Dapp
