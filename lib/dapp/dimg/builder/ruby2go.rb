module Dapp
  module Dimg
    class Builder::Ruby2Go < Builder::Base
      [:before_install, :before_setup, :install, :setup, :build_artifact].each do |stage|
        define_method("#{stage}_checksum") do
          command = "#{snake_case_to_camel_case(stage)}Checksum"
          ruby2go_builder_command(command: command)
        end

        define_method("#{stage}?") do
          command = "Is#{snake_case_to_camel_case(stage)}Empty"
          !ruby2go_builder_command(command: command)
        end

        define_method(stage.to_s) do |image|
          command = snake_case_to_camel_case(stage)
          ruby2go_builder_command(command: command, image: image.ruby2go_image_option).tap do |data|
            image.set_ruby2go_state_hash(JSON.load(data["image"]))
          end
        end
      end

      def ruby2go_builder_command(command:, **options)
        (options[:options] ||= {}).merge!(host_docker_config_dir: dimg.dapp.class.host_docker_config_dir)
        builder = self.class.name.split("::").last.downcase
        command_options = {
          builder: builder,
          command: command,
          config: YAML.dump(dimg.config),
          extra: get_ruby2go_state_hash,
          artifact: dimg.artifact?
        }.merge(options)

        dimg.dapp.ruby2go_builder(command_options).tap do |res|
          raise Error::Build, code: :ruby2go_builder_command_failed_unexpected_error, data: { command: command, message: res["error"] } unless res["error"].nil?
          break res['data']
        end
      end

      def get_ruby2go_state_hash
        {
          "TmpPath" => dimg.tmp_path.to_s,
          "ContainerDappPath" => dimg.container_dapp_path.to_s,
        }
      end

      def snake_case_to_camel_case(value)
        value.to_s.split('_').collect(&:capitalize).join
      end
    end # Builder::Shell
  end # Dimg
end # Dapp
