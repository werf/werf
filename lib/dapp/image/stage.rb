module Dapp
  # Image
  module Image
    # Stage
    class Stage < Docker
      include Argument

      attr_reader :project

      def initialize(name:, project:, built_id: nil, from: nil)
        @project = project

        @container_name = "#{name[/[[^:].]*/]}.#{SecureRandom.hex(4)}"
        @built_id = built_id

        @bash_commands          = []
        @options                = {}
        @change_options         = {}
        @service_change_options = {}

        super(name: name, from: from)
      end

      def built_id
        @built_id ||= id
      end

      def build!(**kwargs)
        run!(**kwargs)
        @built_id = commit!
      ensure
        shellout("docker rm #{container_name}")
      end

      def export!(name, log_verbose: false, log_time: false)
        image = self.class.new(name: name, project: project, built_id: built_id)
        image.tag!(log_verbose: log_verbose, log_time: log_time)
        image.push!(log_verbose: log_verbose, log_time: log_time)
        image.untag!
      end

      def import!(name, log_verbose: false, log_time: false)
        image = self.class.new(name: name, project: project)
        image.pull!(log_verbose: log_verbose, log_time: log_time)
        @built_id = image.built_id
        tag!(log_verbose: log_verbose, log_time: log_time)
        image.untag!
      end

      def tag!(log_verbose: false, log_time: false)
        project.log_warning(desc: { code: :another_image_already_tagged, context: 'warning' }) if !(existed_id = id).nil? && built_id != existed_id
        shellout!("docker tag #{built_id} #{name}", log_verbose: log_verbose, log_time: log_time)
        cache_reset
      end

      protected

      attr_reader :container_name

      def run!(log_verbose: false, log_time: false, introspect_error: false, introspect_before_error: false)
        raise Error::Build, code: :built_id_not_defined if from.built_id.nil?
        shellout!("docker run #{prepared_options} #{from.built_id} -ec '#{prepared_bash_command}'",
                  log_verbose: log_verbose, log_time: log_time)
      rescue Error::Shellout => _e
        raise unless introspect_error || introspect_before_error
        built_id = introspect_error ? commit! : from.built_id
        raise Exception::IntrospectImage, data: { built_id: built_id, options: prepared_options, rmi: introspect_error }
      end

      def commit!
        shellout!("docker commit #{prepared_change} #{container_name}").stdout.strip
      end
    end # Stage
  end # Image
end # Dapp
