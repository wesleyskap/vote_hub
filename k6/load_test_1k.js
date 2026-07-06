import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
  stages: [
    { duration: '30s', target: 1000 }, // Sobe gradualmente até 1.000 VUs
    { duration: '1m', target: 1000 },  // Mantem 1.000 VUs por 1 minuto
    { duration: '30s', target: 1000 }, // Mantem carga em 1.000 VUs
    { duration: '1m', target: 1000 },  // Mantem 1.000 VUs por mais 1 minuto
    { duration: '30s', target: 0 },    // Reduz gradualmente até 0
  ],
  thresholds: {
    http_req_duration: ['p(99)<2000'], // 99% das requisições devem concluir em menos de 2s
  },
};

export default function () {
  const url = __ENV.API_URL || 'http://localhost:30080/api/v1/votes';
  
  // IP aleatório para distribuir carga e simular usuários distintos via X-Forwarded-For
  const randomIP = `${randomIntBetween(1, 255)}.${randomIntBetween(1, 255)}.${randomIntBetween(1, 255)}.${randomIntBetween(1, 255)}`;
  
  const payload = JSON.stringify({
    paredao_id: 1,
    participant_id: randomIntBetween(1, 2), // Juliette ou Gil do Vigor
    fingerprint_id: `k6-test-user-${__VU}-${__ITER}`,
    recaptcha_token: 'test-bypass-token'
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Forwarded-For': randomIP,
    },
  };

  const res = http.post(url, payload, params);
  
  if (res.status !== 202 && res.status !== 429) {
    console.log(`Falhou! Status: ${res.status}, Body: ${res.body}`);
  }

  check(res, {
    'is status 202': (r) => r.status === 202,
    'is status 429': (r) => r.status === 429, // Rastreia respostas de rate limit
  });
  
  sleep(1);
}

export function handleSummary(data) {
  const reqDuration = data.metrics.http_req_duration.values;
  const reqs = data.metrics.http_reqs.values.count;
  const duration = data.state.testRunDurationMs / 1000;
  const rps = (reqs / duration).toFixed(2);
  
  const checkStatus202 = data.metrics.checks.values.passes;
  const totalAttempts = data.metrics.http_reqs.values.count;
  const successPct = totalAttempts > 0 ? ((checkStatus202 / totalAttempts) * 100).toFixed(2) : 0;
  const failedPct = (100 - successPct).toFixed(2);

  return {
    'stdout': `
==================================================
        RELATORIO DE PERFORMANCE - K6 (1k QPS)
==================================================
  Tempo de Execucao:  ${duration.toFixed(1)}s
  Votos Enviados:     ${reqs}
  Vazao Media:        ${rps} req/s
  
  Taxas de Retorno:
    - Sucesso (202):  ${successPct}% (${checkStatus202} votos)
    - Falhas/Outros:  ${failedPct}% (${totalAttempts - checkStatus202} requisicoes)
  
  Tempos de Resposta:
    - Minimo:         ${(reqDuration.min).toFixed(2)}ms
    - Mediana (p50):  ${(reqDuration.med).toFixed(2)}ms
    - P95:            ${(reqDuration['p(95)']).toFixed(2)}ms
    - P99 (SLA):      ${(reqDuration['p(99)']).toFixed(2)}ms
    - Maximo:         ${(reqDuration.max).toFixed(2)}ms
==================================================
`
  };
}
