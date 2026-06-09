class PlayerStat < ApplicationRecord
  belongs_to :user

  validates :user_id, uniqueness: true
  validates :total_score, numericality: { greater_than_or_equal_to: 0 }
  validates :high_score, numericality: { greater_than_or_equal_to: 0 }
  validates :ship_deaths, numericality: { greater_than_or_equal_to: 0 }
  validates :games_played, numericality: { greater_than_or_equal_to: 0 }
  validates :wins, numericality: { greater_than_or_equal_to: 0 }
end
