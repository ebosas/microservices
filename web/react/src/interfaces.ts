export interface Message {
  text: string;
  source: string;
  time: number;
}

export interface MessageCache extends Message {
  timefmt: string;
}

export interface Cache {
  count: number;
  total: number;
  messages: MessageCache[];
}
