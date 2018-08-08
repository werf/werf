module Dapp
  class Dapp
    module Deps
      module Common
        def dappdeps_container(dappdeps)
          dappdeps_containers[dappdeps] ||= ruby2go_dappdeps_command(dappdeps: dappdeps, command: :container)
        end

        def dappdeps_containers
          @dappdeps_containers ||= {}
        end

        def ruby2go_dappdeps_command(dappdeps:, command:, **options)
          ruby2go_dappdeps(dappdeps: dappdeps, command: command, **options).tap do |res|
            unless res["error"].nil?
              raise Error::Dapp, code: :ruby2go_dappdeps_command_failed_unexpected_error,
                                 data: { dappdeps: dappdeps, command: command, message: res["error"] }
            end
            break res['data']
          end
        end
      end # Base
    end # Deps
  end # Dapp
end # Dapp
