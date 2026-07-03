export interface Participant {
  id: number;
  name: string;
  avatar_url?: string;
}

export interface Paredao {
  id: number;
  active: boolean;
  participants: Participant[];
}

export interface VotePayload {
  paredao_id: number;
  participant_id: number;
  recaptcha_token: string;
  fingerprint_id: string;
}

export interface VoteResponse {
  success: boolean;
  message?: string;
}

export interface ParedaoPercentages {
  [participantId: string]: number;
}

export interface ParedaoResults {
  total_votes: number;
  percentages: ParedaoPercentages;
}
