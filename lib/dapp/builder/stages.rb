module Dapp
  module Builder
    module Stages
      STAGES = %i(
        prepare
        infra_install
        sources_1_archive
        sources_1
        app_install
        sources_2
        infra_setup
        sources_3
        app_setup
        sources_4
        sources_5
      ).freeze

      STAGES.reduce(nil) do |dependency, stage|
        define_method :"#{stage}_dependency" do
          dependency
        end

        define_method :"#{stage}_dependency_key" do
          send(:"#{dependency}_key") if dependency
        end

        define_method :"#{stage}_from" do
          send(:"#{dependency}_image_name") unless dependency.nil?
        end

        define_method :"#{stage}_image_name" do
          "dapp:#{send(:"#{stage}_key")}"
        end

        define_method(:"#{stage}_image") do
          instance_variable_get(:"@#{stage}_image") ||
            instance_variable_set(:"@#{stage}_image", Dapp::Image.new(from: send(:"#{stage}_from")))
        end

        define_method :"#{stage}_image_exist?" do
          docker.image_exist?(send("#{stage}_image_name"))
        end

        define_method :"#{stage}_build?" do
          (send(:"#{dependency}_build?") if dependency) or send(:"#{stage}_build_image?")
        end

        define_method :"#{stage}_build_image?" do
          not send(:"#{stage}_image_exist?")
        end

        define_method :"#{stage}_build_image!" do
          image = send(stage)
          docker.build_image!(image: image, name: send(:"#{stage}_image_name")) if image
        end

        define_method :"#{stage}_build!" do
          send(:"#{dependency}_build!") if dependency and send(:"#{dependency}_build?")
          send(:"#{stage}_build_image!") if send(:"#{stage}_build_image?")
        end

        define_method stage do
          begin
            log "   #{stage}"
            opts[:log_indent] += 1

            if block_given?
              yield
            else
              raise
            end
          rescue Exception
            log "=> #{stage} [FAIL]"
          else
            log "=> #{stage} [OK]"
          ensure
            opts[:log_indent] -= 1
          end
        end

        define_method :"#{stage}_key" do
          raise
        end
      end
    end # Stages
  end # Builder
end # Dapp
