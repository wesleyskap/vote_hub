require 'rails_helper'

RSpec.describe "APIs de Votação", type: :request do
  let!(:participant_1) { Participant.create!(name: "Beatriz", avatar_url: "http://img.com/b.jpg") }
  let!(:participant_2) { Participant.create!(name: "Davi", avatar_url: "http://img.com/d.jpg") }

  describe "GET /api/v1/participants" do
    it "deve retornar a lista de participantes cadastrados com sucesso" do
      get "/api/v1/participants"
      expect(response).to have_http_status(:ok)
      
      json = JSON.parse(response.body)
      expect(json.size).to eq(2)
      expect(json.first["name"]).to eq("Beatriz")
    end
  end

  describe "GET /api/v1/results/current" do
    it "deve retornar os resultados acumulados com sucesso" do
      # Adiciona votos agregados
      VoteAggregation.create!(participant: participant_1, total_votes: 150)
      VoteAggregation.create!(participant: participant_2, total_votes: 350)

      get "/api/v1/results/current"
      expect(response).to have_http_status(:ok)

      json = JSON.parse(response.body)
      expect(json).to have_key("total")
      expect(json).to have_key("results")
      expect(json["total"]).to eq(500)
    end
  end

  describe "GET /admin/v1/stats" do
    it "deve retornar estatísticas administrativas do paredão" do
      get "/admin/v1/stats"
      expect(response).to have_http_status(:ok)

      json = JSON.parse(response.body)
      expect(json).to have_key("qps")
      expect(json).to have_key("system_status")
    end
  end
end
