class PlayerMatchResult < ApplicationRecord
  belongs_to :user

  validates :result_id, presence: true, uniqueness: true
  validates :match_id, presence: true
  validates :score, numericality: { only_integer: true, greater_than_or_equal_to: 0 }
  validates :ship_deaths, numericality: { only_integer: true, greater_than_or_equal_to: 0 }
  validates :won, inclusion: { in: [true, false] }
end
