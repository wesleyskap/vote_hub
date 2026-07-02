class Vote < ApplicationRecord
  belongs_to :paredao
  belongs_to :participant

  # Não validamos presence do fingerprint_id pois ele pode vir vazio,
  # a proteção primária no reCAPTCHA já garante que é humano.
end
