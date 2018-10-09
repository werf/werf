require_relative '../spec_helper'

describe Dapp::Dapp::Slug do
  include SpecHelper::Dimg

  it 'a-b-c-d' do
    expect(dapp.consistent_uniq_slugify('a-b-c-d')).to eql('a-b-c-d')
  end

  {
    a_b_c_d:  'a-b-c-d-9736323a',
    a_B_c_d:  'a-b-c-d-a67096ae',
    _a_b_c_d: 'a-b-c-d-72075ca4',
    а_б_ц_д:  'a-b-c-d-fce8fc50',
    "�&�fư&" => 'fu-16a0b26e',
    '------' => '587c09c4',
  }.each do |s, expected_consistent_uniq_slug|
    it s do
      expect(dapp.consistent_uniq_slugify(s.to_s)).to eql(expected_consistent_uniq_slug)
    end
  end

  str = (0..10).map { rand(255).chr }.join
  it "slug(#{str}) == slug(slug(#{str})" do
    consistent_uniq_slug = dapp.consistent_uniq_slugify(str)
    expect(dapp.consistent_uniq_slugify(consistent_uniq_slug)).to eql(consistent_uniq_slug)
  end
end
