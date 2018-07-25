module Dapp
  class Dapp
    module Deps
      module Common
        def dappdeps_container(dappdeps_name)
          dappdeps_containers[dappdeps_name] ||= ruby2go_dappdeps_command(dappdeps_name: dappdeps_name, command: :container)
        end

        def dappdeps_containers
          @dappdeps_containers ||= {}
        end

        def ruby2go_dappdeps_command(dappdeps_name:, command:, **options)
          ruby2go_dappdeps(dappdeps_name: dappdeps_name, command: command, **options).tap do |res|
            unless res["error"].nil?
              raise Error::Dapp, code: :ruby2go_dappdeps_command_failed_unexpected_error,
                                 data: { dappdeps_name: dappdeps_name, command: command, message: res["error"] }
            end
            break res['data']
          end
        end
      end # Base
    end # Deps
  end # Dapp
end # Dapp
