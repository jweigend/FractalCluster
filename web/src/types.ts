export interface CalcParams {
  realMin: number;
  realMax: number;
  imagMin: number;
  imagMax: number;
  picWidth: number;
  picHeight: number;
  blockWidth: number;
  blockHeight: number;
  maxIterations: number;
  maxThreads: number;
  fractalType: string;
  distActive: boolean;
}

export interface BlockStarted {
  type: "block_started";
  blockId: string;
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface BlockResult {
  type: "block_result";
  blockId: string;
  x: number;
  y: number;
  width: number;
  height: number;
  iterations: number[];
}

export interface ProgressUpdate {
  type: "progress";
  completed: number;
  total: number;
}

export interface ErrorMsg {
  type: "error";
  message: string;
}

export type ServerMessage = BlockStarted | BlockResult | ProgressUpdate | ErrorMsg;

export const DEFAULT_PARAMS: CalcParams = {
  realMin: -2.0,
  realMax: 2.0,
  imagMin: -2.0,
  imagMax: 2.0,
  picWidth: 1200,
  picHeight: 800,
  blockWidth: 30,
  blockHeight: 30,
  maxIterations: 100000,
  maxThreads: 100,
  fractalType: "mandelbrot",
  distActive: true,
};
