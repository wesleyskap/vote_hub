$url = "http://localhost:8080/api/v1/votes"

# Usando um ID de participante que não existe (vai gerar um erro de fk constraint no worker (DLQ))
# Enviamos um recaptcha token mockado, assumindo que a API aceitará um mock token se configurada assim
$body = @{
    paredao_id = 1
    participant_id = 999999
    fingerprint_id = "simulate-error"
    recaptcha_token = "valid-token-or-mock"
} | ConvertTo-Json

Invoke-RestMethod -Uri $url -Method Post -Body $body -ContentType "application/json"
