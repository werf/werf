require_relative '../spec_helper'
require_relative 'dimg_spec_helper'

describe Dapp::Dimg::Dimg do
  include SpecHelper::Common
  include SpecHelper::Dimg
  include SpecHelper::Git
  include Integration::DimgSpecHelper

  def project_path
    Pathname('/tmp/dapp/dev_stages')
  end

  def dapp_options
    default_dapp_options.merge(dev: true)
  end

  def change_g_a_post_setup_patch
    file_name = 'large_file'
    if File.exist?(file_name)
      FileUtils.rm(file_name)
    else
      change_file(file_name, random_string(Dapp::Dimg::Build::Stage::Setup::GAPostSetupPatch::MAX_PATCH_SIZE))
    end
  end

  def change_g_a_latest_patch
    change_file
  end

  [:g_a_pre_install_patch, :g_a_post_install_patch, :g_a_pre_setup_patch, :g_a_post_setup_patch, :g_a_latest_patch].each do |stage_name|
    define_method "expect_#{stage_name}_image" do
      check_image_command(stage_name, 'tar -x')
    end
  end

  class_eval(&:generate_stages_tests)
end
