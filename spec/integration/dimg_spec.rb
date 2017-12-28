require_relative '../spec_helper'
require_relative 'dimg_spec_helper'

describe Dapp::Dimg::Dimg do
  include SpecHelper::Common
  include SpecHelper::Dimg
  include SpecHelper::Git
  include Integration::DimgSpecHelper

  def project_path
    Pathname('/tmp/dapp/stages')
  end

  class_eval(&:generate_stages_tests)
end
