module Dapp
  module Build
    module Stage
      # GADependenciesBase
      class GADependenciesBase < Base
        def image
          super do |image|
            application.git_artifacts.each do |git_artifact|
              image.add_service_change_label(git_artifact.full_name.to_sym => git_artifact.latest_commit)
            end
          end
        end

        def empty?
          application.git_artifacts.empty? ? true : false
        end

        protected

        def dependencies_files_checksum(regs)
          unless (files = regs.map { |reg| Dir[File.join(application.home_path, reg)].map { |f| File.read(f) if File.file?(f) } }).empty?
            hashsum files
          end
        end
      end # GADependenciesBase
    end # Stage
  end # Build
end # Dapp
