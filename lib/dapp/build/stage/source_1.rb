module Dapp
  module Build
    module Stage
      class Source1 < SourceBase
        def initialize(build, relative_stage)
          @prev_stage = Source1Archive.new(build, self)
          super
        end

        def name
          :source_1
        end

        def prev_source_stage
          prev_stage
        end

        def signature
          # TODO hashsum [dependencies_checksum, *commit_list]
          hashsum [prev_stage.signature,
                   dependency_file, dependency_file_regex, # FIXME move to git_artifact_signature
                   *build.app_install_commands,
                   *commit_list]
        end

        def git_artifact_signature
          hashsum [prev_stage.signature,
                   *build.app_install_commands]
        end

        private

        def dependency_file
          @dependency_file ||= begin
            file_path = Dir[build.build_path('*')].detect {|x| x =~ dependency_file_regex }
            File.read(file_path) unless file_path.nil?
          end
        end

        def dependency_file?
          !dependency_file.nil?
        end

        def dependency_file_regex
          /.*\/(Gemfile|composer.json|requirement_file.txt)$/
        end
      end # Source1
    end # Stage
  end # Build
end # Dapp
