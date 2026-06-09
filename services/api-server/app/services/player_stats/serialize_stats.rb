module PlayerStats
  class SerializeStats
    def self.call(player_stat:)
      {
        total_score: player_stat.total_score,
        high_score: player_stat.high_score,
        ship_deaths: player_stat.ship_deaths,
        games_played: player_stat.games_played,
        wins: player_stat.wins
      }
    end
  end
end
