class Participant < ApplicationRecord
  has_many :paredao_participants, dependent: :destroy
  has_many :paredaos, through: :paredao_participants
  has_many :votes
  has_many :vote_aggregations_by_hours

  validates :name, presence: true
end
