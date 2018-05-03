class CreateUser < ActiveRecord::Migration[5.2]
  def change
    create_table :users, id: false do |t|
      t.string :uuid
      t.integer :level
      t.integer :exp
      t.string :name
      t.boolean :online
    end

    add_index :users, :uuid, unique: true
  end
end
