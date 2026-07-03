module Admin
  module V1
    class StatsController < ApplicationController
      # GET /admin/api/v1/stats
      # Retorna métricas analíticas agregadas para o painel de controle administrativo da produção
      # TODO: Melhorar essas tratativas
      def index
        paredao = Paredao.find_by(status: 'active')
        if paredao.nil?
          return render json: { error: 'Nenhum paredão ativo encontrado' }, status: :not_found
        end

        # Total geral de votos do paredão ativo
        total_votes = VoteAggregationsByHour.where(paredao_id: paredao.id).sum(:total_votes)

        # Divisão de votos por nome do participante
        aggregations = VoteAggregationsByHour.where(paredao_id: paredao.id).group(:participant_id).sum(:total_votes)
        votes_by_participant = {}
        
        paredao.participants.each do |p|
          votes_by_participant[p.name] = aggregations[p.id] || 0
        end

        # Histórico de votos agregados por hora (últimas 24 horas)
        hourly_data = VoteAggregationsByHour.where(paredao_id: paredao.id)
                                            .where("vote_hour >= ?", 24.hours.ago)
                                            .order(:vote_hour)
                                            .group(:vote_hour)
                                            .sum(:total_votes)

        votes_by_hour = hourly_data.map do |hour, votes|
          {
            hour: hour.strftime("%Y-%m-%d %H:%M"),
            votes: votes
          }
        end

        render json: {
          total_votes: total_votes,
          votes_by_participant: votes_by_participant,
          votes_by_hour: votes_by_hour
        }
      end
    end
  end
end
