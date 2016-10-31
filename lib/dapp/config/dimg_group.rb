module Dapp
  module Config
    class DimgGroup < DimgGroupBase
      include Dimg::InstanceMethods

      attr_reader :_artifacts

      def initialize(project:)
        @_artifacts ||= []
        super
      end

      def dimg(_name = nil, &_blk)
        super.tap do |dimg|
          dimg._chef          = _chef
          dimg._shell         = _shell
          dimg._docker        = _docker
          dimg._git_artifacts = _git_artifacts
          dimg._mounts        = _mounts
        end
      end

      def artifact(&blk)
        check_dimg_group_directive_order
        _artifacts << Directive::Artifact.new(project: _project, &blk)
      end

      def chef(&blk)
        check_dimg_directive_order
        super
      end

      def shell(&blk)
        check_dimg_directive_order
        super
      end

      def docker(&blk)
        check_dimg_directive_order
        super
      end

      def git_artifact(type_or_git_repo, &blk)
        check_dimg_directive_order
        super
      end

      def mount(to, &blk)
        check_dimg_directive_order
        super
      end

      protected

      def check_dimg_directive_order
        raise if check_dimg_group_directive_order or _artifacts.any? # TODO: adding order
      end

      def check_dimg_group_directive_order
        raise if _dimgs.any? or _dimgs_groups.any? # TODO: adding order
      end
    end
  end
end
