# Limpando banco antes de popular
VoteAggregationsByHour.delete_all
Vote.delete_all
ParedaoParticipant.delete_all
Paredao.delete_all
Participant.delete_all

puts "Plantando dados..."

# 1. Criação dos Participantes do BBB
p1 = Participant.create!(
  name: "Alane",
  avatar_url: "https://example.com/avatars/alane.jpg"
)

p2 = Participant.create!(
  name: "Davi",
  avatar_url: "https://example.com/avatars/davi.jpg"
)

# 2. Criação do Paredão
paredao = Paredao.create!(status: "active")

# 3. Associação: Quem está no Paredão
ParedaoParticipant.create!(paredao: paredao, participant: p1)
ParedaoParticipant.create!(paredao: paredao, participant: p2)

puts "Banco populado com sucesso!"
puts "ID Paredão Atual: #{paredao.id}"
puts "ID #{p1.name}: #{p1.id}"
puts "ID #{p2.name}: #{p2.id}"
