module Dapp
  module Builder
    module Stages
      STAGES_DEPENDENCIES = {
          prepare: nil,
          infra_install: :prepare,
          sources_1: :infra_install,
          infra_setup: :sources_1,
          app_install: :infra_setup,
          sources_2: :app_install,
          app_setup: :sources_2,
          sources_3: :app_setup,
          #sources_4: :sources_3
      }.freeze

      STAGES_DEPENDENCIES.each do |stage, dependence|
        define_method :"#{stage}_dependence" do
          dependence
        end

        define_method :"#{stage}_dependence_key" do
          send(:"#{dependence}_key") if dependence
        end

        define_method :"#{stage}_from" do
          send(:"#{dependence}_image_name") unless dependence.nil?
        end

        define_method :"#{stage}_image_name" do
          "dapp:#{send(:"#{stage}_key")}"
        end

        define_method(:"#{stage}_image") do
          instance_variable_get(:"@#{stage}_image") ||
            instance_variable_set(:"@#{stage}_image", Image.new(from: send(:"#{stage}_from")))
        end

        define_method :"#{stage}_image_exist?" do
          docker.image_exist?(send("#{stage}_image_name"))
        end

        define_method :"#{stage}_build?" do
          (send(:"#{dependence}_build?") if dependence) or send(:"#{stage}_build_image?")
        end

        define_method :"#{stage}_build_image?" do
          not send(:"#{stage}_image_exist?")
        end

        define_method :"#{stage}_build_image!" do
          image = send(stage)
          docker.build_image!(image: image, name: send(:"#{stage}_image_name")) if image
        end

        define_method :"#{stage}_build!" do
          send(:"#{dependence}_build!") if dependence and send(:"#{dependence}_build?")
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
