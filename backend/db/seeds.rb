# Execução de seed segura e concorrente usando advisory locks do PostgreSQL e restart de identidade
Participant.transaction do
  # Obtém uma trava transacional (advisory lock) para garantir que apenas um pod execute por vez
  ActiveRecord::Base.connection.execute("SELECT pg_advisory_xact_lock(742159)")

  if Participant.count == 0
    puts "Inserindo dados..."

    # Trunca as tabelas e reinicia as sequências (identities) a partir de 1 de forma atômica
    ActiveRecord::Base.connection.execute("TRUNCATE TABLE votes, vote_aggregations_by_hours, paredao_participants, paredaos, participants RESTART IDENTITY CASCADE")

    # 1. Criação dos Participantes do BBB (IDs serão gerados sequencialmente a partir de 1)
    p1 = Participant.create!(
      name: "Alane",
      avatar_url: "https://example.com/avatars/alane.jpg"
    )

    p2 = Participant.create!(
      name: "Davi",
      avatar_url: "https://example.com/avatars/davi.jpg"
    )

    # 2. Criação do Paredão (ID será gerado sequencialmente a partir de 1)
    paredao = Paredao.create!(status: "active")

    # 3. Associação: Quem está no Paredão
    ParedaoParticipant.create!(paredao: paredao, participant: p1)
    ParedaoParticipant.create!(paredao: paredao, participant: p2)

    puts "Banco populado com sucesso!"
    puts "ID Paredão Atual: #{paredao.id}"
    puts "ID #{p1.name}: #{p1.id}"
    puts "ID #{p2.name}: #{p2.id}"
  else
    puts "Participantes já existentes no banco de dados. Pulando plantio."
  end
end


