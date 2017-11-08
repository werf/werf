module Dapp
  module Kube
    class Helm::Release
      include Helper::YAML

      attr_reader :dapp

      attr_reader :name
      attr_reader :repo
      attr_reader :image_version
      attr_reader :namespace
      attr_reader :chart_path
      attr_reader :set
      attr_reader :values
      attr_reader :deploy_timeout

      def initialize(dapp,
        name:, repo:, image_version:, namespace:, chart_path:,
        set: [], values: [], deploy_timeout: nil)
        @dapp = dapp

        @name = name
        @repo = repo
        @image_version = image_version
        @namespace = namespace
        @chart_path = chart_path
        @set = set
        @values = values
        @deploy_timeout = deploy_timeout
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

      def deploy!
        args = [
          name, chart_path, additional_values_options,
          set_options, upgrade_extra_options
        ].flatten

        dapp.kubernetes.create_namespace!(namespace) unless dapp.kubernetes.namespace?(namespace)

        dapp.shellout! "helm upgrade #{args.join(' ')}", verbose: true
      end

      def templates
        @templates ||= {}.tap do |t|
          current_template = nil
          spec = 0
          evaluation_output.lines.each do |l|
            if (match = l[/# Source: #{dapp.name}\/templates\/(.*)/, 1])
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
            t[template] = "---\n#{specs.map(&:join).join("---\n").strip}"
          end
        end
      end

      protected

      def evaluation_output
        @evaluation_output ||= begin
          cmd = dapp.shellout! [
            "helm",
            "template",
            chart_path,
            additional_values_options,
            set_options,
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

      def additional_values_options
        [].tap do |options|
          options.concat(values.map { |p| "--values #{p}" })
        end
      end

      def set_options
        [].tap do |options|
          options << "--set global.dapp.repo=#{repo}"
          options << "--set global.dapp.image_version=#{image_version}"
          options << "--set global.namespace=#{namespace}"
          options.concat(set.map { |opt| "--set #{opt}" })
        end
      end

      def upgrade_extra_options(dry_run: nil)
        dry_run = dapp.dry_run? if dry_run.nil?

        [].tap do |options|
          options << "--namespace #{namespace}" if namespace
          options << '--install'
          options << '--dry-run' if dry_run
          options << '--debug'   if dry_run
          options << "--timeout #{deploy_timeout}" if deploy_timeout
        end
      end
    end # Helm::Release
  end # Kube
end # Dapp
