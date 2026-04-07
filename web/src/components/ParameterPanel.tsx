import type { CalcParams } from "../types";

interface Props {
  params: CalcParams;
  onParamsChange: (params: CalcParams) => void;
  onStart: () => void;
  onStop: () => void;
  calculating: boolean;
  connected: boolean;
  showBlockOutlines: boolean;
  onToggleBlockOutlines: () => void;
}

export function ParameterPanel({
  params,
  onParamsChange,
  onStart,
  onStop,
  calculating,
  connected,
  showBlockOutlines,
  onToggleBlockOutlines,
}: Props) {
  const update = (key: keyof CalcParams, value: number | string | boolean) => {
    onParamsChange({ ...params, [key]: value });
  };

  const numField = (label: string, key: keyof CalcParams, step?: number) => (
    <label style={styles.field}>
      <span style={styles.label}>{label}</span>
      <input
        type="number"
        value={params[key] as number}
        step={step ?? (typeof params[key] === "number" && (params[key] as number) % 1 !== 0 ? 0.01 : 1)}
        onChange={(e) => update(key, parseFloat(e.target.value))}
        style={styles.input}
      />
    </label>
  );

  return (
    <div style={styles.panel}>
      <h3 style={styles.heading}>Coordinates</h3>
      {numField("Real Min", "realMin", 0.01)}
      {numField("Real Max", "realMax", 0.01)}
      {numField("Imag Min", "imagMin", 0.01)}
      {numField("Imag Max", "imagMax", 0.01)}

      <h3 style={styles.heading}>Image</h3>
      {numField("Width (px)", "picWidth")}
      {numField("Height (px)", "picHeight")}

      <h3 style={styles.heading}>Computation</h3>
      {numField("Max Iterations", "maxIterations")}
      {numField("Max Threads", "maxThreads")}

      <h3 style={styles.heading}>Distribution</h3>
      <label style={styles.field}>
        <span style={styles.label}>Enabled</span>
        <input
          type="checkbox"
          checked={params.distActive}
          onChange={(e) => update("distActive", e.target.checked)}
        />
      </label>
      {params.distActive && (
        <>
          {numField("Block Width", "blockWidth")}
          {numField("Block Height", "blockHeight")}
        </>
      )}

      <h3 style={styles.heading}>Display</h3>
      <label style={styles.field}>
        <span style={styles.label}>Block Outlines</span>
        <input
          type="checkbox"
          checked={showBlockOutlines}
          onChange={onToggleBlockOutlines}
        />
      </label>

      <div style={styles.buttons}>
        {calculating ? (
          <button
            onClick={onStop}
            style={{ ...styles.btn, ...styles.btnStop }}
          >
            Stop
          </button>
        ) : (
          <button
            onClick={onStart}
            disabled={!connected}
            style={{
              ...styles.btn,
              ...styles.btnStart,
              ...(!connected ? styles.btnDisabled : {}),
            }}
          >
            Start
          </button>
        )}
      </div>

      <div style={{ ...styles.status, color: connected ? "#4a4" : "#a44" }}>
        {connected ? "Connected" : "Disconnected"}
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  panel: {
    width: 240,
    padding: "12px 16px",
    background: "#1a1a2e",
    borderRight: "1px solid #333",
    overflowY: "auto",
    fontSize: 13,
    color: "#ccc",
  },
  heading: {
    margin: "14px 0 6px",
    fontSize: 12,
    textTransform: "uppercase",
    color: "#888",
    letterSpacing: 1,
  },
  field: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 4,
  },
  label: { flex: 1 },
  input: {
    width: 90,
    padding: "2px 4px",
    background: "#0d0d1a",
    border: "1px solid #444",
    color: "#eee",
    borderRadius: 3,
    textAlign: "right" as const,
  },
  buttons: {
    display: "flex",
    gap: 8,
    marginTop: 16,
  },
  btn: {
    flex: 1,
    padding: "6px 0",
    border: "none",
    borderRadius: 4,
    cursor: "pointer",
    fontWeight: "bold",
    fontSize: 13,
  },
  btnStart: { background: "#2a6", color: "#fff", cursor: "pointer" },
  btnStop: { background: "#a44", color: "#fff", cursor: "pointer" },
  btnDisabled: { opacity: 0.4, cursor: "not-allowed" },
  status: {
    marginTop: 8,
    fontSize: 11,
    textAlign: "center" as const,
  },
};
