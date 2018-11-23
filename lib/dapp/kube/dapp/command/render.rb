module Dapp
  module Kube
    module Dapp
      module Command
        module Render
          def kube_render
            command = "render"

            # TODO: move project dir logic to golang
            project_dir = path.to_s

            res = ruby2go_deploy(
              "command" => command,
              "projectDir" => project_dir,
              "rubyCliOptions" => JSON.dump(self.options),
            )

            raise ::Dapp::Error::Command, code: :ruby2go_deploy_command_failed, data: { command: command, message: res["error"] } unless res["error"].nil?
          end

          def kube_render_old
            helm_release do |release|
              templates = begin
                if options[:templates].any?
                  release.templates.select do |template_path, _|
                    options[:templates].map { |t| "#{t}*" }.any? do |template_path_pattern|
                      template_path_without_chart_name = template_path[/.*?\/(.*)/, 1]
                      template_relative_path_pattern = Pathname(File.expand_path(template_path_pattern)).subpath_of(path('.helm'))
                      template_relative_path_pattern ||= template_path_pattern

                      File.fnmatch?(template_relative_path_pattern, template_path_without_chart_name, File::FNM_PATHNAME|File::FNM_DOTMATCH) ||
                        File.fnmatch?(template_relative_path_pattern, template_path, File::FNM_PATHNAME|File::FNM_DOTMATCH)
                    end
                  end
                else
                  release.templates
                end
              end

              templates.values.each { |t| puts t }
            end
          end
        end
      end
    end
  end
end
