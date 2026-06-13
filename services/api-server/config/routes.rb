Rails.application.routes.draw do
  # Define your application routes per the DSL in https://guides.rubyonrails.org/routing.html

  # Reveal health status on /up that returns 200 if the app boots with no exceptions, otherwise 500.
  # Can be used by load balancers and uptime monitors to verify that the app is live.
  get "up" => "rails/health#show", as: :rails_health_check
  get "health" => "health#show"

  namespace :api do
    namespace :auth do
      post "register", to: "registrations#create"
      post "login", to: "sessions#create"
      delete "logout", to: "sessions#destroy"
      get "me", to: "me#show"
      get "discord/start", to: "discord#start"
      get "discord/callback", to: "discord#callback"
      post "discord/login_sessions", to: "discord_login_sessions#create"
      post "discord/login_sessions/:id/exchange", to: "discord_login_sessions#exchange"
    end

    namespace :internal do
      namespace :player_data, path: "player-data" do
        post "stats", to: "stats#create"
      end
    end

    namespace :player do
      get "stats", to: "stats#show"
    end
  end

  namespace :internal do
    namespace :auth do
      post "verify-token", to: "verify_tokens#create"
    end

    namespace :player_data, path: "player-data" do
      post "match-results", to: "match_results#create"
    end
  end

  # Defines the root path route ("/")
  # root "posts#index"
end
