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

ActiveRecord::Schema[8.1].define(version: 2026_06_08_000700) do
  # These are extensions that must be enabled in order to support this database
  enable_extension "pg_catalog.plpgsql"

  create_table "access_tokens", force: :cascade do |t|
    t.string "audience", default: "api", null: false
    t.datetime "created_at", null: false
    t.datetime "expires_at", null: false
    t.datetime "last_used_at"
    t.datetime "revoked_at"
    t.string "token_digest", null: false
    t.datetime "updated_at", null: false
    t.bigint "user_id", null: false
    t.index ["audience"], name: "index_access_tokens_on_audience"
    t.index ["expires_at"], name: "index_access_tokens_on_expires_at"
    t.index ["revoked_at"], name: "index_access_tokens_on_revoked_at"
    t.index ["token_digest"], name: "index_access_tokens_on_token_digest", unique: true
    t.index ["user_id"], name: "index_access_tokens_on_user_id"
  end

  create_table "oauth_login_sessions", force: :cascade do |t|
    t.datetime "consumed_at"
    t.datetime "created_at", null: false
    t.datetime "expires_at", null: false
    t.string "poll_secret_digest", null: false
    t.string "provider", null: false
    t.string "public_id", null: false
    t.string "status", null: false
    t.datetime "updated_at", null: false
    t.bigint "user_id"
    t.index ["expires_at"], name: "index_oauth_login_sessions_on_expires_at"
    t.index ["poll_secret_digest"], name: "index_oauth_login_sessions_on_poll_secret_digest", unique: true
    t.index ["public_id"], name: "index_oauth_login_sessions_on_public_id", unique: true
    t.index ["user_id"], name: "index_oauth_login_sessions_on_user_id"
  end

  create_table "oauth_states", force: :cascade do |t|
    t.datetime "consumed_at"
    t.datetime "created_at", null: false
    t.datetime "expires_at", null: false
    t.bigint "oauth_login_session_id"
    t.string "provider", null: false
    t.string "redirect_after"
    t.string "state_digest", null: false
    t.datetime "updated_at", null: false
    t.index ["expires_at"], name: "index_oauth_states_on_expires_at"
    t.index ["oauth_login_session_id"], name: "index_oauth_states_on_oauth_login_session_id"
    t.index ["provider"], name: "index_oauth_states_on_provider"
    t.index ["state_digest"], name: "index_oauth_states_on_state_digest", unique: true
  end

  create_table "password_credentials", force: :cascade do |t|
    t.datetime "created_at", null: false
    t.string "email", null: false
    t.string "password_digest", null: false
    t.datetime "updated_at", null: false
    t.bigint "user_id", null: false
    t.index ["email"], name: "index_password_credentials_on_email", unique: true
    t.index ["user_id"], name: "index_password_credentials_on_user_id", unique: true
  end

  create_table "user_identities", force: :cascade do |t|
    t.datetime "created_at", null: false
    t.string "email"
    t.string "provider", null: false
    t.string "provider_uid", null: false
    t.datetime "updated_at", null: false
    t.bigint "user_id", null: false
    t.index ["email"], name: "index_user_identities_on_email"
    t.index ["provider", "provider_uid"], name: "index_user_identities_on_provider_and_provider_uid", unique: true
    t.index ["user_id"], name: "index_user_identities_on_user_id"
  end

  create_table "users", force: :cascade do |t|
    t.datetime "created_at", null: false
    t.string "display_name", null: false
    t.datetime "updated_at", null: false
  end

  add_foreign_key "access_tokens", "users"
  add_foreign_key "oauth_login_sessions", "users"
  add_foreign_key "oauth_states", "oauth_login_sessions"
  add_foreign_key "password_credentials", "users"
  add_foreign_key "user_identities", "users"
end
