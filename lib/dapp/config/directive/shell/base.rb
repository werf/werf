module Dapp
  module Config
    module Directive
      module Shell
        class Base
          attr_reader :_before_install, :_before_setup, :_install, :_setup
          attr_reader :_before_install_cache_version, :_before_setup_cache_version, :_install_cache_version, :_setup_cache_version

          def initialize
            @_before_install = []
            @_before_setup   = []
            @_install        = []
            @_setup          = []
          end

          def before_install(*args, cache_version: nil)
            @_before_install.concat(args)
            @_before_install_cache_version = cache_version
          end

          def before_setup(*args, cache_version: nil)
            @_before_setup.concat(args)
            @_before_setup_cache_version = cache_version
          end

          def install(*args, cache_version: nil)
            _install.concat(args)
            @_install_cache_version = cache_version
          end

          def setup(*args, cache_version: nil)
            _setup.concat(args)
            @_setup_cache_version = cache_version
          end

          def reset_before_install
            @_before_install = []
          end

          def reset_before_setup
            @_before_setup = []
          end

          def reset_install
            @_install = []
          end

          def reset_setup
            @_setup = []
          end

          def reset_all
            methods.tap { |arr| arr.delete(__method__) }.grep(/^reset_/).each(&method(:send))
          end

          protected

          def clone
            marshal_dup(self)
          end

          def clone_to_artifact
            Artifact.new.tap do |shell|
              self.instance_variables.each do |variable|
                shell.instance_variable_set(variable, marshal_dup(self.instance_variable_get(variable)))
                shell.instance_variable_set("#{variable}_cache_version", marshal_dup(self.instance_variable_get("#{variable}_cache_version")))
              end
            end
          end

          def empty?
            @_before_install.empty? && @_before_setup.empty? && @_install.empty? && @_setup.empty?
          end

          private

          def marshal_dup(obj)
            Marshal.load(Marshal.dump(obj))
          end
        end
      end
    end
  end
end
