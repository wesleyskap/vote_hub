require 'rails_helper'

RSpec.describe Participant, type: :model do
  describe 'Validações' do
    it 'deve ser válido com nome presente (Fast & Independent)' do
      participant = Participant.new(name: 'Beatriz', avatar_url: 'http://img.com/b.jpg')
      expect(participant).to be_valid
    end

    it 'deve ser inválido sem nome' do
      participant = Participant.new(name: nil)
      expect(participant).not_to be_valid
      expect(participant.errors[:name]).to include("can't be blank")
    end
  end

  describe 'Relacionamentos' do
    it 'deve poder participar de vários paredões' do
      assoc = described_class.reflect_on_association(:paredaos)
      expect(assoc.macro).to eq :has_many
    end
  end
end
