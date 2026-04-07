import type { CalcParams, ServerMessage } from "../types";

export class FractalSocket {
  private ws: WebSocket | null = null;
  private onMessage: (msg: ServerMessage) => void;
  private onStatusChange: (connected: boolean) => void;

  constructor(
    onMessage: (msg: ServerMessage) => void,
    onStatusChange: (connected: boolean) => void
  ) {
    this.onMessage = onMessage;
    this.onStatusChange = onStatusChange;
  }

  connect() {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      this.onStatusChange(true);
    };

    this.ws.onclose = () => {
      this.onStatusChange(false);
    };

    this.ws.onmessage = (event) => {
      const msg = JSON.parse(event.data) as ServerMessage;
      this.onMessage(msg);
    };
  }

  startCalculation(params: CalcParams) {
    this.send({ type: "start", params });
  }

  stopCalculation() {
    this.send({ type: "stop" });
  }

  private send(data: unknown) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  disconnect() {
    this.ws?.close();
    this.ws = null;
  }
}
