module Jekyll
  module Variables
    def increment_shared_counter(name)
        @context.registers[:werf_docs_variables] ||= {}
        @context.registers[:werf_docs_variables][name] ||= 0

        old_value = @context.registers[:werf_docs_variables][name]
        @context.registers[:werf_docs_variables][name] += 1

        return old_value
    end

    def reset_shared_counter(name)
        @context.registers[:werf_docs_variables] ||= {}
        @context.registers[:werf_docs_variables][name] = 0
        return
    end
  end
end

Liquid::Template.register_filter(Jekyll::Variables)
