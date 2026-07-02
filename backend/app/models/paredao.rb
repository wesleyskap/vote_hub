class Paredao < ApplicationRecord
  has_many :paredao_participants, dependent: :destroy
  has_many :participants, through: :paredao_participants
  has_many :votes
  has_many :vote_aggregations_by_hours

  validates :status, presence: true, inclusion: { in: %w[active closed] }
end
