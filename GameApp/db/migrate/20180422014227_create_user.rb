class CreateUser < ActiveRecord::Migration[5.2]
  def change
    create_table :users do |t|
      t.string :users
      t.string :uuid
      t.integer :level
      t.integer :exp
      t.string :name
      t.boolean :online
    end
  end
end
