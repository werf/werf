module Dapp
  module Deployment
    module Config
      module Directive
        class App < Base
          module InstanceMethods
            attr_reader :_dimg, :_namespace, :_expose, :_bootstrap, :_migrate, :_run

            def self.included(base)
              base.include(Namespace::InstanceMethods)
            end

            def dimg(name)
              @_dimg = name
            end

            def namespace(name, &blk)
              (_namespace[name] ||= begin
                Namespace.new(name, dapp: dapp).tap do |namespace|
                  _namespace[name] = namespace
                end
              end).tap do |namespace|
                directive_eval(namespace, &blk)
              end
            end

            def expose(&blk)
              directive_eval(_expose, &blk)
            end

            def bootstrap(cmd)
              sub_directive_eval { @_bootstrap = cmd }
            end

            def migrate(cmd)
              sub_directive_eval { @_migrate = cmd }
            end

            def run(cmd)
              sub_directive_eval { @_run = cmd }
            end

            protected

            def app_init_variables!
              @_migrate   = []
              @_namespace = {}
              @_expose    = Expose.new(dapp: dapp)
            end

            def passed_directives
              [:@_dimg, :@_namespace, :@_expose, :@_bootstrap, :@_migrate,
               :@_run, :@_environment, :@_secret_environment, :@_scale]
            end
          end # InstanceMethods
        end # App
      end # Directive
    end # Config
  end # Deployment
end # Dapp
