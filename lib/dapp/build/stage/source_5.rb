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

        def dependencies_stage
          nil
        end

        def dependencies
          [commit_list, change_options]
        end

        def image
          super do |image|
            change_options.each do |k, v|
              next if v.nil? || v.empty?
              image.public_send("add_change_#{k}", v)
            end
          end
        end

        def layer_commit(git_artifact)
          commits[git_artifact] ||= begin
            git_artifact.latest_commit
          end
        end

        def empty?
          dependencies_empty?
        end

        private

        def change_options
          application.config._docker._change_options
        end

        def commit_list
          application.git_artifacts.map { |git_artifact| layer_commit(git_artifact) }
        end
      end # Source5
    end # Stage
  end # Build
end # Dapp
