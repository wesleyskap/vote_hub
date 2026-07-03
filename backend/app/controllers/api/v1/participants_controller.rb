module Api
  module V1
    class ParticipantsController < ApplicationController
      # GET /api/v1/participants
      # Retorna a lista de participantes do paredão ativo para a interface do frontend
      def index
        paredao = Paredao.find_by(status: 'active')
        if paredao.nil?
          return render json: { error: 'Nenhum paredão ativo encontrado' }, status: :not_found
        end

        # Seleciona apenas os campos necessários alinhados à interface do TypeScript
        participants = paredao.participants.order(:id)
        render json: participants.as_json(only: [:id, :name, :avatar_url])
      end
    end
  end
end
