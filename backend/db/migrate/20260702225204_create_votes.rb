class CreateVotes < ActiveRecord::Migration[8.1]
  def change
    create_table :votes do |t|
      t.references :paredao, null: false, foreign_key: true
      t.references :participant, null: false, foreign_key: true
      t.string :fingerprint_id
      t.datetime :created_at, null: false, default: -> { 'CURRENT_TIMESTAMP' }
    end
    
    # BRIN index na coluna created_at garantindo performance para milhões de inserções
    add_index :votes, :created_at, using: :brin
  end
end
