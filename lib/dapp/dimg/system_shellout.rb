module Dapp
  # Dimg
  class Dimg
    # SystemShellout
    module SystemShellout
      def system_shellout(command, **kwargs)
        project.system_shellout_extra(volume: tmp_path) do
          project.system_shellout(command, **kwargs)
        end
      end

      def system_shellout!(command, **kwargs)
        system_shellout(command, raise_error: true, **kwargs)
      end
    end # SystemShellout
  end # Dimg
end # Dapp
