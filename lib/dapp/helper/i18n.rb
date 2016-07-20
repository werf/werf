module Dapp
  module Helper
    # I18n
    module I18n
      def i18n_initialize
        ::I18n.load_path << Dir[File.join(Dapp.root, 'config', '**', '*')].select { |path| File.file?(path) }
        ::I18n.reload!
        ::I18n.locale = :en
      end

      def t(desc: {}, context: nil)
        code = desc[:code]
        data = desc[:data]
        ::I18n.t [:common, context, code].join('.'), [:common, code].join('.'), **data, raise: true
      rescue ::I18n::MissingTranslationData => _e
        raise NetStatus::Exception, code: :missing_translation, data: { code: code }
      end
    end
  end # Helper
end # Dapp
