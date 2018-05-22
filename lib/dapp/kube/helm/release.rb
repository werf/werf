module Dapp
  module Kube
    class Helm::Release
      include Helper::YAML

      attr_reader :dapp

      attr_reader :name
      attr_reader :repo
      attr_reader :docker_tag
      attr_reader :namespace
      attr_reader :chart_path
      attr_reader :set
      attr_reader :values
      attr_reader :deploy_timeout
      attr_reader :without_registry

      def initialize(dapp,
        name:, repo:, docker_tag:, namespace:, chart_path:,
        set: [], values: [], deploy_timeout: nil, without_registry: nil)
        @dapp = dapp

        @name = name
        @repo = repo
        @docker_tag = docker_tag
        @namespace = namespace
        @chart_path = chart_path
        @set = set
        @values = values
        @deploy_timeout = deploy_timeout
        @without_registry = (without_registry.nil? ? false : without_registry)
      end

      def jobs
        (resources_specs['Job'] || {}).map do |name, spec|
          [name, Kubernetes::Client::Resource::Job.new(spec)]
        end.to_h
      end

      def hooks
        jobs.select do |_, spec|
          spec.annotations.key? "helm.sh/hook"
        end
      end

      def deployments
        (resources_specs['Deployment'] || {}).map do |name, spec|
          [name, Kubernetes::Client::Resource::Deployment.new(spec)]
        end.to_h
      end

      def install_helm_release
        unless dapp.dry_run?
          dapp.kubernetes.create_namespace!(namespace) unless dapp.kubernetes.namespace?(namespace)
        end

        cmd = dapp.shellout([
          "helm install #{chart_path}",
          "--name #{name}",
          *helm_additional_values_options,
          *helm_set_options,
          *helm_install_options,
        ].join(" "))

        return cmd
      end

      def upgrade_helm_release
        cmd = dapp.shellout([
          "helm upgrade #{name} #{chart_path}",
          *helm_additional_values_options,
          *helm_set_options,
          *helm_install_options
        ].join(" "))

        return cmd
      end

      def templates
        @templates ||= {}.tap do |t|
          current_template = nil
          spec = 0
          evaluation_output.lines.each do |l|
            if (match = l[/# Source: (.*)/, 1])
              spec = 0
              t[current_template = match] ||= []
            end

            if l[/^---$/]
              spec += 1
            elsif current_template
              (t[current_template][spec] ||= []) << l
            end
          end

          t.each do |template, specs|
            t[template] = "---\n#{specs.reject(&:nil?).map(&:join).join("---\n").strip}"
          end
        end
      end

      def lint!
        dapp.shellout! [
          'helm',
          'lint',
          '--strict',
          *helm_additional_values_options,
          *helm_set_options(fake: true),
          *helm_common_options,
          chart_path
        ].compact.join(' ')
      end

      protected

      def evaluation_output
        @evaluation_output ||= begin
          cmd = dapp.shellout! [
            "helm",
            "template",
            chart_path,
            helm_additional_values_options,
            helm_set_options(without_registry: true),
            ("--namespace #{namespace}" if namespace),
          ].compact.join(" ")

          cmd.stdout
        end
      end

      def resources_specs
        @resources_specs ||= {}.tap do |specs|
          evaluation_output.split(/^---$/)
              .reject {|chunk| chunk.lines.all? {|line| line.strip.empty? or line.strip.start_with? "#"}}
              .map {|chunk| yaml_load(chunk)}.each do |spec|
            specs[spec['kind']] ||= {}
            specs[spec['kind']][(spec['metadata'] || {})['name']] = spec
          end
        end
      end

      def helm_additional_values_options
        [].tap do |options|
          options.concat(values.map { |p| "--values #{p}" })
        end
      end

      def dimg_registry
        @dimg_registry ||= dapp.dimg_registry(repo)
      end

      def helm_set_options(without_registry: false, fake: false)
        [].tap do |options|
          options.concat set.map {|opt| "--set #{opt}"}

          service_values = Helm::Values.service_values(dapp, repo, namespace, docker_tag,
                                                       without_registry: self.without_registry || without_registry,
                                                       fake: fake)
          options.concat service_values.to_set_options
        end
      end

      def helm_install_options(dry_run: nil)
        dry_run = dapp.dry_run? if dry_run.nil?

        helm_common_options(dry_run: dry_run).tap do |options|
          options << '--dry-run' if dry_run
          options << "--timeout #{deploy_timeout}" if deploy_timeout
        end
      end

      def helm_common_options(dry_run: nil)
        dry_run = dapp.dry_run? if dry_run.nil?

        [].tap do |options|
          options << "--namespace #{namespace}" if namespace
          options << '--debug'                  if dry_run
        end
      end
    end # Helm::Release
  end # Kube
end # Dapp
