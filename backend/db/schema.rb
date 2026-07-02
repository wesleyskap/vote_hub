# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema[8.1].define(version: 2026_07_02_225205) do
  # These are extensions that must be enabled in order to support this database
  enable_extension "pg_catalog.plpgsql"

  create_table "paredao_participants", force: :cascade do |t|
    t.datetime "created_at", null: false
    t.bigint "paredao_id", null: false
    t.bigint "participant_id", null: false
    t.datetime "updated_at", null: false
    t.index ["paredao_id", "participant_id"], name: "index_paredao_participants_on_paredao_id_and_participant_id", unique: true
    t.index ["paredao_id"], name: "index_paredao_participants_on_paredao_id"
    t.index ["participant_id"], name: "index_paredao_participants_on_participant_id"
  end

  create_table "paredaos", force: :cascade do |t|
    t.datetime "created_at", null: false
    t.string "status", default: "active", null: false
    t.datetime "updated_at", null: false
  end

  create_table "participants", force: :cascade do |t|
    t.string "avatar_url"
    t.datetime "created_at", null: false
    t.string "name", null: false
    t.datetime "updated_at", null: false
  end

  create_table "vote_aggregations_by_hours", primary_key: ["paredao_id", "participant_id", "vote_hour"], force: :cascade do |t|
    t.integer "paredao_id", null: false
    t.integer "participant_id", null: false
    t.bigint "total_votes", default: 0, null: false
    t.datetime "vote_hour", null: false
  end

  create_table "votes", force: :cascade do |t|
    t.datetime "created_at", default: -> { "CURRENT_TIMESTAMP" }, null: false
    t.string "fingerprint_id"
    t.bigint "paredao_id", null: false
    t.bigint "participant_id", null: false
    t.index ["created_at"], name: "index_votes_on_created_at", using: :brin
    t.index ["paredao_id"], name: "index_votes_on_paredao_id"
    t.index ["participant_id"], name: "index_votes_on_participant_id"
  end

  add_foreign_key "paredao_participants", "paredaos"
  add_foreign_key "paredao_participants", "participants"
  add_foreign_key "vote_aggregations_by_hours", "paredaos"
  add_foreign_key "vote_aggregations_by_hours", "participants"
  add_foreign_key "votes", "paredaos"
  add_foreign_key "votes", "participants"
end
