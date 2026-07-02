class CreateParedaos < ActiveRecord::Migration[8.1]
  def change
    create_table :paredaos do |t|
      t.string :status, null: false, default: 'active'
      t.timestamps
    end
  end
end
