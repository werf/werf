module Dapp
  module Dimg
    module Image
      class Stage < Docker
        include Argument

        def initialize(name:, dapp:, built_id: nil, from: nil)
          @container_name = "#{dapp.name}.#{SecureRandom.hex(4)}"
          @built_id = built_id

          @bash_commands          = []
          @service_bash_commands  = []
          @options                = {}
          @change_options         = {}
          @service_change_options = {}

          super(name: name, dapp: dapp, from: from)
        end

        def built_id_example
          # FIXME: remove this example
          self.dapp.ruby2go_image(Hash[instance_variables.map { |name| [name, instance_variable_get(name).to_s] } ])
        end

        def built_id
          @built_id ||= id
        end

        def build!
          run!
          @built_id = commit!
        ensure
          dapp.shellout("#{dapp.host_docker} rm #{container_name}")
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
        end

        def labels
          config_option('Labels')
        end

        protected

        attr_reader :container_name

        def run!
          raise Error::Build, code: :built_id_not_defined if from.built_id.nil?
          dapp.shellout!("#{dapp.host_docker} run #{prepared_options} #{from.built_id} -ec '#{prepared_bash_command}'", verbose: true)
        rescue ::Dapp::Error::Shellout => error
          dapp.log_warning(desc: { code: :launched_command, data: { command: prepared_commands.join(' && ') }, context: :container })

          raise unless dapp.introspect_error? || dapp.introspect_before_error?
          built_id = dapp.introspect_error? ? commit! : from.built_id
          raise Exception::IntrospectImage, data: { built_id: built_id,
                                                    options: prepared_options,
                                                    rmi: dapp.introspect_error?,
                                                    error: error }
        end

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
