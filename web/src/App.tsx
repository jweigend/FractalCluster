import { useState, useRef, useCallback, useEffect } from "react";
import type { CalcParams, BlockStarted, BlockResult, ServerMessage } from "./types";
import { DEFAULT_PARAMS } from "./types";
import { FractalSocket } from "./lib/websocket";
import { FractalRenderer } from "./lib/renderer";
import { FractalCanvas } from "./components/FractalCanvas";
import { ParameterPanel } from "./components/ParameterPanel";
import { StatusBar } from "./components/StatusBar";

function App() {
  const [params, setParams] = useState<CalcParams>(DEFAULT_PARAMS);
  const [calculating, setCalculating] = useState(false);
  const [connected, setConnected] = useState(false);
  const [completed, setCompleted] = useState(0);
  const [total, setTotal] = useState(0);
  const [showBlockOutlines, setShowBlockOutlines] = useState(true);

  const rendererRef = useRef<FractalRenderer | null>(null);
  const socketRef = useRef<FractalSocket | null>(null);
  const showBlockOutlinesRef = useRef(showBlockOutlines);
  showBlockOutlinesRef.current = showBlockOutlines;

  const handleMessage = useCallback(
    (msg: ServerMessage) => {
      switch (msg.type) {
        case "block_started": {
          if (showBlockOutlinesRef.current) {
            const block = msg as BlockStarted;
            rendererRef.current?.drawBlockOutline(block.x, block.y, block.width, block.height);
          }
          break;
        }
        case "block_result": {
          const block = msg as BlockResult;
          rendererRef.current?.renderBlock(block);
          break;
        }
        case "progress":
          setCompleted(msg.completed);
          setTotal(msg.total);
          if (msg.completed >= msg.total) {
            setCalculating(false);
          }
          break;
        case "error":
          console.error("Server error:", msg.message);
          break;
      }
    },
    []
  );

  useEffect(() => {
    const socket = new FractalSocket(handleMessage, setConnected);
    socket.connect();
    socketRef.current = socket;
    return () => socket.disconnect();
  }, [handleMessage]);

  const handleStart = useCallback((overrideParams?: CalcParams) => {
    const p = overrideParams ?? params;
    rendererRef.current?.clear();
    rendererRef.current?.updateColorTable(p.maxIterations);
    setCalculating(true);
    setCompleted(0);
    setTotal(0);
    socketRef.current?.startCalculation(p);
  }, [params]);

  const handleStop = useCallback(() => {
    socketRef.current?.stopCalculation();
    setCalculating(false);
  }, []);

  return (
    <div style={styles.app}>
      <ParameterPanel
        params={params}
        onParamsChange={setParams}
        onStart={() => handleStart()}
        onStop={handleStop}
        calculating={calculating}
        connected={connected}
        showBlockOutlines={showBlockOutlines}
        onToggleBlockOutlines={() => setShowBlockOutlines(!showBlockOutlines)}
      />
      <div style={styles.main}>
        <div style={styles.canvasWrap}>
          <FractalCanvas
            params={params}
            onParamsChange={setParams}
            onStartCalculation={handleStart}
            rendererRef={rendererRef}
          />
        </div>
        <StatusBar completed={completed} total={total} calculating={calculating} />
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  app: {
    display: "flex",
    height: "100vh",
    background: "#0d0d1a",
    color: "#eee",
    fontFamily: "'Segoe UI', system-ui, sans-serif",
  },
  main: {
    flex: 1,
    display: "flex",
    flexDirection: "column",
    minWidth: 0,
  },
  canvasWrap: {
    flex: 1,
    display: "flex",
    justifyContent: "center",
    alignItems: "center",
    overflow: "auto",
    padding: 16,
  },
};

export default App;
