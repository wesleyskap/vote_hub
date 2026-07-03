Rails.application.routes.draw do
  # Define your application routes per the DSL in https://guides.rubyonrails.org/routing.html

  # Reveal health status on /up that returns 200 if the app boots with no exceptions, otherwise 500.
  get "up" => "rails/health#show", as: :rails_health_check

  namespace :api do
    namespace :v1 do
      resources :participants, only: [:index]
      get "results/current", to: "results#current"
    end
  end

  namespace :admin do
    namespace :v1 do
      get "stats", to: "stats#index"
    end
  end
end
