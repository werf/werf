module Dapp
  module Dimg
    module Image
      class Stage < Docker
        include Argument

        def initialize(name:, dapp:, built_id: nil, from: nil)
          @built_id = built_id

          @bash_commands          = []
          @service_bash_commands  = []
          @options                = {}
          @change_options         = {}
          @service_change_options = {}

          super(name: name, dapp: dapp, from: from)
        end

        def labels
          built_image_inspect!['Config']['Labels']
        end

        def created_at
          built_image_inspect!["Created"]
        end

        def size
          Float(built_image_inspect!["Size"])
        end

        def built_image_inspect
          @built_image_inspect || image_inspect
        end

        def built_image_inspect!
          raise Error::Build, code: :image_not_exist, data: { name: name } unless built?
          built_image_inspect
        end

        def build_options
          [
            :name,
            :from,
            :built_image_inspect,
            :image_inspect,
            :built_id,
            :bash_commands,
            :service_bash_commands,
            :options,
            :change_options,
            :service_change_options,
          ].map do |name|
            if name == :from && !from.nil?
              [name, from.build_options] unless from.nil?
            else
              [name, send(name)]
            end
          end
            .compact
            .to_h
        end

        def build!
          res = self.dapp.ruby2go_image(**ruby2go_image_build_options)
          if res["error"].nil?
            image = JSON.load(res['data']['image'])
            @built_id = image['built_id']
            @built_image_inspect = image['built_image_inspect']
          elsif res["error"].start_with? "stage build failed: container run failed"
            raise Error::Build, code: :ruby2go_image_command_failed, data: { command: "build" }
          else
            raise Error::Build, code: :ruby2go_image_command_failed_unexpected_error, data: { command: "build", message: res["error"] }
          end
        end

        def introspect!
          res = self.dapp.ruby2go_image(**ruby2go_image_introspect_options)
          raise Error::Build, code: :ruby2go_image_command_failed_unexpected_error, data: { command: "introspect", message: res["error"] } unless res["error"].nil?
        end

        def ruby2go_image_build_options
          {
            image: JSON.dump(build_options),
            command: "build",
            options: {
              introspection: {
                before: dapp.introspect_before_error?,
                after: dapp.introspect_error?
              }
            }
          }
        end

        def ruby2go_image_introspect_options
          {
            image: JSON.dump(build_options),
            command: "introspect",
          }
        end

        def built_id
          @built_id ||= id
        end

        def built?
          !built_id.nil?
        end

        def export!(name)
          tag!(name).tap do |image|
            image.push!
            image.untag!
          end
        end

        def tag!(name)
          clone!(name).tap do |image|
            self.class.tag!(id: image.built_id, tag: image.name)
            image.reset_image_inspect
          end
        end

        def import!(name)
          clone!(name).tap do |image|
            image.pull!
            @built_id = image.built_id
            save_in_cache!
            image.untag!
          end
        end

        def save_in_cache!
          dapp.log_warning(desc: { code: :another_image_already_tagged }) if !(existed_id = id).nil? && built_id != existed_id
          self.class.tag!(id: built_id, tag: name)
          reset_image_inspect
        end

        protected

        def clone!(name)
          self.class.new(name: name, dapp: dapp, built_id: built_id)
        end
      end # Stage
    end # Image
  end # Dimg
end # Dapp
