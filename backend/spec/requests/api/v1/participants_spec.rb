require 'rails_helper'

RSpec.describe "APIs de Votacao", type: :request do
  before(:all) do
    # Limpa dados pre-existentes (como de seeds) para garantir independencia nos testes
    VoteAggregationsByHour.delete_all
    ParedaoParticipant.delete_all
    Participant.delete_all
    Paredao.delete_all
  end

  let!(:participant_1) { Participant.create!(name: "Beatriz", avatar_url: "http://img.com/b.jpg") }
  let!(:participant_2) { Participant.create!(name: "Davi", avatar_url: "http://img.com/d.jpg") }

  describe "GET /api/v1/participants" do
    it "deve retornar a lista de participantes cadastrados com sucesso" do
      get "/api/v1/participants"
      expect(response).to have_http_status(:ok)
      
      json = JSON.parse(response.body)
      expect(json.size).to eq(2)
      expect(json.map { |p| p["name"] }).to include("Beatriz", "Davi")
    end
  end

  describe "GET /api/v1/results/current" do
    it "deve retornar os resultados acumulados com sucesso" do
      paredao = Paredao.create!(status: "active")
      ParedaoParticipant.create!(paredao: paredao, participant: participant_1)
      ParedaoParticipant.create!(paredao: paredao, participant: participant_2)

      # Adiciona votos agregados por hora
      VoteAggregationsByHour.create!(paredao: paredao, participant: participant_1, total_votes: 150, vote_hour: Time.current.beginning_of_hour)
      VoteAggregationsByHour.create!(paredao: paredao, participant: participant_2, total_votes: 350, vote_hour: Time.current.beginning_of_hour)

      get "/api/v1/results/current"
      expect(response).to have_http_status(:ok)

      json = JSON.parse(response.body)
      expect(json).to have_key("total_votes")
      expect(json).to have_key("percentages")
      expect(json["total_votes"]).to eq(500)
    end
  end

  describe "GET /admin/v1/stats" do
    it "deve retornar estatisticas administrativas do paredao" do
      paredao = Paredao.create!(status: "active")
      ParedaoParticipant.create!(paredao: paredao, participant: participant_1)
      ParedaoParticipant.create!(paredao: paredao, participant: participant_2)

      get "/admin/v1/stats"
      expect(response).to have_http_status(:ok)

      json = JSON.parse(response.body)
      expect(json).to have_key("total_votes")
      expect(json).to have_key("votes_by_participant")
      expect(json).to have_key("votes_by_hour")
    end
  end
end
