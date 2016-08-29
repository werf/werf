# Prctl
module Prctl
  PR_SET_PDEATHSIG = 1

  class << self
    def _fiddle_func
      require 'fiddle'
      @_fiddle_func ||= Fiddle::Function.new(
        Fiddle::Handle['prctl'.freeze], [
          Fiddle::TYPE_INT,
          Fiddle::TYPE_LONG,
          Fiddle::TYPE_LONG,
          Fiddle::TYPE_LONG,
          Fiddle::TYPE_LONG
        ], Fiddle::TYPE_INT
      )
    end

    def call(*a)
      _fiddle_func.call(*a)
    end
  end
end
