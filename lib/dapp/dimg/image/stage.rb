module Dapp
  module Dimg
    module Image
      class Stage < Docker
        include Argument

        def initialize(name:, dapp:, built_id: nil, from: nil)
          @container_name = "dapp.build.#{SecureRandom.hex(4)}"
          @built_id = built_id

          @bash_commands          = []
          @service_bash_commands  = []
          @options                = {}
          @change_options         = {}

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
            :built_id,
            :container_name,
            # :bash_commands, // TODO
            # :service_bash_commands, // TODO
            :prepared_bash_command,
            :options,
            :service_options,
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
          res = self.dapp.ruby2go_image(image: JSON.dump(build_options), command: "build")
          if res["error"].nil?
            image = JSON.load(res['data']['image'])
            @built_id = image['built_id']
            @built_image_inspect = image['built_image_inspect']
          else
            begin
              if res["error"].start_with?("container run failed")
                dapp.log_warning(desc: { code: :launched_command, data: { command: prepared_commands.join(' && ') }, context: :container })

                error_build_error = Error::Build.new(code: :ruby2go_image_command_failed, data: { command: "build" })
                raise error_build_error unless dapp.introspect_error? || dapp.introspect_before_error?
                built_id = dapp.introspect_error? ? commit! : from.built_id
                raise Exception::IntrospectImage, data: { built_id: built_id,
                                                          options: prepared_options,
                                                          rmi: dapp.introspect_error?,
                                                          error: error_build_error }
              else
                raise res["error"]
              end
            ensure
              dapp.shellout("#{dapp.host_docker} rm #{container_name}") # TODO
            end
          end
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

        attr_reader :container_name

        def commit!
          dapp.shellout!("#{dapp.host_docker} commit #{prepared_change} #{container_name}").stdout.strip
        end

        def clone!(name)
          self.class.new(name: name, dapp: dapp, built_id: built_id)
        end
      end # Stage
    end # Image
  end # Dimg
end # Dapp
