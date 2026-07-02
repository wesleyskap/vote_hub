package vote

// Payload representa a estrutura do voto enviado pelo frontend.
// Seus campos estão alinhados por tamanho em bytes (maior -> menor)
// para otimização de padding do compilador.
type Payload struct {
	Fingerprint    string `json:"fingerprint_id"`    // 16 bytes (ptr + len)
	RecaptchaToken string `json:"recaptcha_token"`   // 16 bytes (ptr + len)
	ParedaoID      int64  `json:"paredao_id"`        // 8 bytes
	ParticipantID  int64  `json:"participant_id"`    // 8 bytes
}
