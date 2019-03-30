class CreatePlayerData < ActiveRecord::Migration[5.2]
  def change
    create_table :player_datas, id: false do |t|
      t.string :uuid
      t.text :content, limit: 16777215
      t.integer :updated_at
    end

    add_index :player_datas, :uuid, unique: true
    add_index :player_datas, :updated_at
  end
end
