module Api
  module V1
    class ResultsController < ApplicationController
      # GET /api/v1/results/current
      # Retorna o total consolidado de votos e os percentuais parciais de cada candidato em tempo real
      def current
        paredao = Paredao.find_by(status: 'active')
        if paredao.nil?
          return render json: { error: 'Nenhum paredão ativo encontrado' }, status: :not_found
        end

        # Consulta otimizada agregando por participante a partir da tabela consolidada por hora
        aggregations = VoteAggregationsByHour.where(paredao_id: paredao.id).group(:participant_id).sum(:total_votes)
        
        participants_ids = paredao.participants.pluck(:id)
        total_votes = 0
        votes_map = {}

        # Mapeia os votos garantindo que mesmo participantes sem votos fiquem com 0
        participants_ids.each do |pid|
          votes = aggregations[pid] || 0
          votes_map[pid.to_s] = votes
          total_votes += votes
        end

        # Calcula o percentual arredondado com duas casas decimais
        percentages = {}
        votes_map.each do |pid, votes|
          percentages[pid] = total_votes > 0 ? ((votes.to_f / total_votes) * 100).round(2) : 0.0
        end

        render json: {
          total_votes: total_votes,
          percentages: percentages
        }
      end
    end
  end
end
