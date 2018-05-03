class CreateEquip < ActiveRecord::Migration[5.2]
  def change
    create_table :equips, id: false do |t|
      t.string :uuid
      t.string :user_id
      t.integer :level
      t.integer :conf_id
      t.string :evolves
      t.string :equips
      t.integer :exp
    end

    add_index :equips, :uuid, unique: true
  end
end
