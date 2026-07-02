class CreateParticipants < ActiveRecord::Migration[8.1]
  def change
    create_table :participants do |t|
      t.string :name, null: false
      t.string :avatar_url
      t.timestamps
    end
  end
end
