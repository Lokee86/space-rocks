module PlayerStats
  class ApplyMatchResult
    Result = Struct.new(:player_stat, :match_result, :duplicate, :error, keyword_init: true) do
      def success?
        error.nil?
      end
    end

    def self.call(user:, result_id:, match_id:, score:, ship_deaths:, won:)
      if user.nil? || result_id.nil? || result_id == "" || match_id.nil? || match_id == ""
        return Result.new(error: :invalid_input)
      end

      player_stat = nil
      match_result = nil

      ActiveRecord::Base.transaction do
        existing_match_result = PlayerMatchResult.find_by(result_id: result_id)
        if existing_match_result
          return Result.new(
            player_stat: user.player_stat,
            match_result: existing_match_result,
            duplicate: true
          )
        end

        player_stat = user.player_stat || user.create_player_stat!
        match_result = PlayerMatchResult.create!(
          result_id: result_id,
          match_id: match_id,
          user: user,
          score: score,
          ship_deaths: ship_deaths,
          won: won
        )

        player_stat.update!(
          games_played: player_stat.games_played + 1,
          total_score: player_stat.total_score + score,
          high_score: [player_stat.high_score, score].max,
          ship_deaths: player_stat.ship_deaths + ship_deaths,
          wins: player_stat.wins + (won ? 1 : 0)
        )
      end

      Result.new(
        player_stat: player_stat,
        match_result: match_result,
        duplicate: false
      )
    end
  end
end
