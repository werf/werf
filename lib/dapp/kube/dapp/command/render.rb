module Dapp
  module Kube
    module Dapp
      module Command
        module Render
          def kube_render
            helm_release { |release| release.templates.values.each { |t| puts t } }
          end
        end
      end
    end
  end
end
