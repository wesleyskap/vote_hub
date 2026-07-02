class CreateVoteAggregationsByHours < ActiveRecord::Migration[8.1]
  def change
    create_table :vote_aggregations_by_hours, id: false do |t|
      t.integer :paredao_id, null: false
      t.integer :participant_id, null: false
      t.datetime :vote_hour, null: false
      t.bigint :total_votes, default: 0, null: false
    end

    # Criando a chave primária composta (Covering Index O(log N))
    execute <<-SQL
      ALTER TABLE vote_aggregations_by_hours
      ADD PRIMARY KEY (paredao_id, participant_id, vote_hour);
    SQL

    add_foreign_key :vote_aggregations_by_hours, :paredaos
    add_foreign_key :vote_aggregations_by_hours, :participants
  end
end
