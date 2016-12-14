module Dapp
  module Config
    class Dimg < Base
      # Merging
      module Merging
        protected

        [:mount, :artifact].each do |directive|
          define_method "merge_#{directive}" do |a, b|
            b.map { |a| a.send(:_clone) }.concat(Array(a))
          end
        end

        [:install_dependencies, :setup_dependencies, :artifact_dependencies].each do |directive|
          define_method "merge_#{directive}" do |a, b|
            b.dup.concat(Array(a))
          end
        end

        [:builder, :dev_mode].each do |directive|
          define_method "merge_#{directive}" do |a, b|
            a.nil? ? b : a
          end
        end
      end
    end
  end
end
