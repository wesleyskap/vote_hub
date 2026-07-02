class CreateParedaoParticipants < ActiveRecord::Migration[8.1]
  def change
    create_table :paredao_participants do |t|
      t.references :paredao, null: false, foreign_key: true
      t.references :participant, null: false, foreign_key: true

      t.timestamps
    end
    add_index :paredao_participants, [:paredao_id, :participant_id], unique: true
  end
end
