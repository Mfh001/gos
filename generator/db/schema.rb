# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# Note that this schema.rb definition is the authoritative source for your
# database schema. If you need to create the application database on another
# system, you should be using db:schema:load, not running all the migrations
# from scratch. The latter is a flawed and unsustainable approach (the more migrations
# you'll amass, the slower it'll run and the greater likelihood for issues).
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema.define(version: 2019_03_29_163827) do

  create_table "equips", id: false, options: "ENGINE=InnoDB DEFAULT CHARSET=utf8", force: :cascade do |t|
    t.string "uuid"
    t.string "user_id"
    t.integer "level"
    t.integer "conf_id"
    t.string "evolves"
    t.string "equips"
    t.integer "exp"
    t.index ["uuid"], name: "index_equips_on_uuid", unique: true
  end

  create_table "player_datas", id: false, options: "ENGINE=InnoDB DEFAULT CHARSET=utf8", force: :cascade do |t|
    t.string "uuid"
    t.text "content", limit: 16777215
    t.integer "updated_at"
    t.index ["uuid"], name: "index_player_datas_on_uuid", unique: true
  end

  create_table "schema_persistances", id: false, options: "ENGINE=InnoDB DEFAULT CHARSET=utf8", force: :cascade do |t|
    t.string "uuid"
    t.integer "version"
    t.index ["uuid"], name: "index_schema_persistances_on_uuid", unique: true
  end

  create_table "users", id: false, options: "ENGINE=InnoDB DEFAULT CHARSET=utf8", force: :cascade do |t|
    t.string "uuid"
    t.integer "level"
    t.integer "exp"
    t.string "name"
    t.boolean "online"
    t.index ["uuid"], name: "index_users_on_uuid", unique: true
  end

end
