import type { BlockResult } from "../types";
import { buildColorTable } from "./colorTable";

export class FractalRenderer {
  private ctx: CanvasRenderingContext2D;
  private colorTable: Uint32Array;

  constructor(ctx: CanvasRenderingContext2D, maxIterations: number) {
    this.ctx = ctx;
    this.colorTable = buildColorTable(maxIterations);
  }

  updateColorTable(maxIterations: number, colorChange?: number, colorOffset?: number) {
    this.colorTable = buildColorTable(maxIterations, colorChange, colorOffset);
  }

  clear() {
    this.ctx.fillStyle = "#000";
    this.ctx.fillRect(0, 0, this.ctx.canvas.width, this.ctx.canvas.height);
  }

  renderBlock(block: BlockResult) {
    const { x, y, width, height, iterations } = block;
    const imageData = this.ctx.createImageData(width, height);
    const buf = new Uint32Array(imageData.data.buffer);

    for (let i = 0; i < iterations.length; i++) {
      const iter = iterations[i];
      buf[i] = this.colorTable[Math.min(iter, this.colorTable.length - 1)];
    }

    this.ctx.putImageData(imageData, x, y);
  }

  drawBlockOutline(x: number, y: number, width: number, height: number) {
    this.ctx.strokeStyle = "red";
    this.ctx.lineWidth = 1;
    this.ctx.strokeRect(x + 0.5, y + 0.5, width - 1, height - 1);
  }
}
