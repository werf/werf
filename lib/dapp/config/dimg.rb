module Dapp
  module Config
    # Dimg
    class Dimg < Base
      include Validation
      include InstanceMethods

      attr_reader :_name

      def initialize(name, dapp:)
        self._name = name
        super(dapp: dapp)
      end

      def _name=(name)
        return if name.nil?
        reg = '^[[[:alnum:]]_.-]*$'
        raise Error::Config, code: :dimg_name_incorrect, data: { name: name, reg: reg } unless name =~ /#{reg}/
        @_name = name
      end
    end
  end
end
