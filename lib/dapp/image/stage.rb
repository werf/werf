module Dapp
  # Image
  module Image
    # Stage
    class Stage < Docker
      def initialize(name:, built_id: nil, from: nil)
        @bash_commands = []
        @options = {}
        @change_options = {}
        @container_name = SecureRandom.hex
        @built_id = built_id
        super(name: name, from: from)
      end

      def built_id
        @built_id ||= id
      end

      def build!(**kwargs)
        @built_id = if should_be_built?
                      begin
                        run!(**kwargs)
                        commit!
                      ensure
                        shellout("docker rm #{container_name}")
                      end
                    else
                      from.built_id
                    end
      end

      def export!(name, log_verbose: false, log_time: false, force: false)
        image = self.class.new(built_id: built_id, name: name)
        image.tag!(log_verbose: log_verbose, log_time: log_time, force: force)
        image.push!(log_verbose: log_verbose, log_time: log_time)
        image.untag!
      end

      def tag!(log_verbose: false, log_time: false, force: false)
        if !(existed_id = id).nil? && !force
          raise Error::Build, code: :another_image_already_tagged if built_id != existed_id
          return
        end
        shellout!("docker tag #{built_id} #{name}", log_verbose: log_verbose, log_time: log_time)
      end

      protected

      attr_reader :container_name

      def run!(log_verbose: false, log_time: false, introspect_error: false, introspect_before_error: false)
        raise Error::Build, code: :built_id_not_defined if from.built_id.nil?
        shellout!("docker run #{prepared_options} --name=#{container_name} #{from.built_id} #{prepared_bash_command}",
                  log_verbose: log_verbose, log_time: log_time)
      rescue Error::Shellout => e
        raise unless introspect_error || introspect_before_error
        built_id = introspect_error ? commit! : from.built_id
        raise Exception::IntrospectImage, message: Dapp::Helper::NetStatus.message(e),
                                          data: { built_id: built_id, options: prepared_options, rmi: introspect_error }
      end

      def commit!
        shellout!("docker commit #{prepared_change} #{container_name}").stdout.strip
      end

      def should_be_built?
        !(bash_commands.empty? && change_options.empty?)
      end
    end
  end # Image
end # Dapp
