class VoteAggregationsByHour < ApplicationRecord
  # Informando o ActiveRecord da tabela explícita sem convenção de plural do rails (pois terminamos em by_hours)
  self.table_name = "vote_aggregations_by_hours"
  
  belongs_to :paredao
  belongs_to :participant
end
