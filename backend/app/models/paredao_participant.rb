class ParedaoParticipant < ApplicationRecord
  belongs_to :paredao
  belongs_to :participant

  validates :paredao_id, uniqueness: { scope: :participant_id }
end
