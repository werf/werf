module Dapp
  class Dapp
    module Logging
      # I18n
      module I18n
        def self.initialize
          ::I18n.load_path << Dir[File.join(::Dapp.root, 'config', '**', '*')].select { |path| File.file?(path) }
          ::I18n.reload!
          ::I18n.locale = :en
        end

        def t(context: nil, **desc)
          code = desc[:code]
          data = desc[:data] || {}
          paths = []
          paths << [:common, context, code].join('.') if context
          paths << [:common, code].join('.')
          ::I18n.t(*paths, **data, raise: true)
        rescue ::I18n::MissingTranslationData => _e
          raise ::NetStatus::Exception, code: :missing_translation, data: { code: code }
        end
      end
    end # Helper
  end
end # Dapp
