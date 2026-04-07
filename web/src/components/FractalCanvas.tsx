import { useRef, useEffect, useCallback, useState } from "react";
import type { CalcParams } from "../types";
import { FractalRenderer } from "../lib/renderer";

interface Props {
  params: CalcParams;
  onParamsChange: (params: CalcParams) => void;
  onStartCalculation: (overrideParams?: CalcParams) => void;
  rendererRef: React.RefObject<FractalRenderer | null>;
}

export function FractalCanvas({
  params,
  onParamsChange,
  onStartCalculation,
  rendererRef,
}: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const overlayRef = useRef<HTMLCanvasElement>(null);
  const [dragStart, setDragStart] = useState<{ x: number; y: number; button: number } | null>(null);
  const [dragEnd, setDragEnd] = useState<{ x: number; y: number } | null>(null);

  // Initialize renderer when canvas or maxIterations changes
  useEffect(() => {
    const canvas = canvasRef.current;
    const overlay = overlayRef.current;
    if (!canvas || !overlay) return;
    canvas.width = params.picWidth;
    canvas.height = params.picHeight;
    overlay.width = params.picWidth;
    overlay.height = params.picHeight;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    rendererRef.current = new FractalRenderer(ctx, params.maxIterations);
    rendererRef.current.clear();
  }, [params.picWidth, params.picHeight, params.maxIterations, rendererRef]);

  const getCanvasPos = useCallback(
    (e: React.MouseEvent<HTMLCanvasElement>) => {
      const overlay = overlayRef.current!;
      const rect = overlay.getBoundingClientRect();
      const scaleX = overlay.width / rect.width;
      const scaleY = overlay.height / rect.height;
      return {
        x: (e.clientX - rect.left) * scaleX,
        y: (e.clientY - rect.top) * scaleY,
      };
    },
    []
  );

  const handleMouseDown = useCallback(
    (e: React.MouseEvent<HTMLCanvasElement>) => {
      e.preventDefault();
      const pos = getCanvasPos(e);
      setDragStart({ ...pos, button: e.button });
      setDragEnd(null);
    },
    [getCanvasPos]
  );

  const handleMouseMove = useCallback(
    (e: React.MouseEvent<HTMLCanvasElement>) => {
      if (!dragStart) return;
      const pos = getCanvasPos(e);
      // Maintain aspect ratio
      const dx = pos.x - dragStart.x;
      const aspect = params.picHeight / params.picWidth;
      const dy = Math.abs(dx) * aspect * Math.sign(pos.y - dragStart.y);
      setDragEnd({ x: dragStart.x + dx, y: dragStart.y + dy });
    },
    [dragStart, getCanvasPos, params.picWidth, params.picHeight]
  );

  const handleMouseUp = useCallback(
    (_e: React.MouseEvent<HTMLCanvasElement>) => {
      if (!dragStart || !dragEnd) {
        setDragStart(null);
        return;
      }

      const x1 = Math.min(dragStart.x, dragEnd.x);
      const x2 = Math.max(dragStart.x, dragEnd.x);
      const y1 = Math.min(dragStart.y, dragEnd.y);
      const y2 = Math.max(dragStart.y, dragEnd.y);

      if (x2 - x1 < 2) {
        setDragStart(null);
        setDragEnd(null);
        return;
      }

      const { realMin, realMax, imagMin, imagMax, picWidth, picHeight } = params;
      const rScale = (realMax - realMin) / picWidth;
      const iScale = (imagMax - imagMin) / picHeight;

      const isRightClick = dragStart.button === 2;

      let newParams: CalcParams;
      if (isRightClick) {
        // Zoom out: interpret box as visible portion, expand beyond
        const bw = x2 - x1;
        const bh = y2 - y1;
        newParams = {
          ...params,
          realMin: realMin - (realMax - realMin) / bw * x1,
          realMax: realMax + (realMax - realMin) / bw * (picWidth - x2),
          imagMax: imagMax + (imagMax - imagMin) / bh * y1,
          imagMin: imagMin - (imagMax - imagMin) / bh * (picHeight - y2),
        };
      } else {
        // Zoom in: map box to new coordinates
        newParams = {
          ...params,
          realMin: realMin + x1 * rScale,
          realMax: realMin + x2 * rScale,
          imagMax: imagMax - y1 * iScale,
          imagMin: imagMax - y2 * iScale,
        };
      }

      setDragStart(null);
      setDragEnd(null);
      onParamsChange(newParams);
      onStartCalculation(newParams);
    },
    [dragStart, dragEnd, params, onParamsChange, onStartCalculation]
  );

  const handleContextMenu = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
  }, []);

  // Draw zoom box on overlay canvas (keeps fractal canvas clean)
  useEffect(() => {
    const overlay = overlayRef.current;
    if (!overlay) return;
    const ctx = overlay.getContext("2d");
    if (!ctx) return;

    ctx.clearRect(0, 0, overlay.width, overlay.height);

    if (!dragStart || !dragEnd) return;

    const x = Math.min(dragStart.x, dragEnd.x);
    const y = Math.min(dragStart.y, dragEnd.y);
    const w = Math.abs(dragEnd.x - dragStart.x);
    const h = Math.abs(dragEnd.y - dragStart.y);

    ctx.strokeStyle = "white";
    ctx.lineWidth = 1;
    ctx.setLineDash([4, 4]);
    ctx.strokeRect(x, y, w, h);
    ctx.setLineDash([]);
  });

  return (
    <div style={{ position: "relative", display: "inline-block", maxWidth: "100%" }}>
      <canvas
        ref={canvasRef}
        style={{
          border: "1px solid #333",
          display: "block",
          maxWidth: "100%",
          height: "auto",
        }}
      />
      <canvas
        ref={overlayRef}
        style={{
          position: "absolute",
          top: 0,
          left: 0,
          width: "100%",
          height: "100%",
          cursor: "crosshair",
        }}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onContextMenu={handleContextMenu}
      />
    </div>
  );
}
